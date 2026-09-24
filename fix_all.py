import os
import re

# 1. GitHub Actions: deploy.yml
os.makedirs('.github/workflows', exist_ok=True)
with open('.github/workflows/deploy.yml', 'w') as f:
    f.write("""name: Build & Deploy

on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    permissions:
      id-token: write
      contents: read
    steps:
      - uses: actions/checkout@v4

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: arn:aws:iam::YOUR_ACCOUNT:role/github-actions-ecr-push
          aws-region: us-east-1

      - name: Login to ECR
        uses: aws-actions/amazon-ecr-login@v2

      - name: Build & push
        run: |
          # Replace YOUR_ECR_REPO with actual repository URI in production
          ECR_REPO="YOUR_ECR_REPO"
          docker build -f infra/Dockerfile \\
            -t $ECR_REPO:${{ github.sha }} \\
            -t $ECR_REPO:latest .
          # Uncomment below when ECR is ready
          # docker push $ECR_REPO:${{ github.sha }}
          # docker push $ECR_REPO:latest
""")

# 2. GitHub Actions: go-test.yml
test_yml_path = '.github/workflows/go-test.yml'
if os.path.exists(test_yml_path):
    with open(test_yml_path, 'r') as f:
        test_yml = f.read()
    if 'integration:' not in test_yml:
        test_yml += """
  integration:
    runs-on: ubuntu-latest
    needs: test
    services:
      postgres:
        image: postgres:17
        env:
          POSTGRES_PASSWORD: test
          POSTGRES_DB: strykz_test
        options: >-
          --health-cmd pg_isready
          --health-interval 5s
      redis:
        image: redis:7
        options: --health-cmd "redis-cli ping"
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.27"
      - run: go test -count=1 -tags=integration ./...
        env:
          POSTGRES_URL: postgres://postgres:test@localhost:5432/strykz_test
          REDIS_ADDRESS: localhost:6379
"""
        with open(test_yml_path, 'w') as f:
            f.write(test_yml)

# 3. Dockerfile
dockerfile_path = 'infra/Dockerfile'
with open(dockerfile_path, 'w') as f:
    f.write("""# syntax=docker/dockerfile:1

# Build The Application From Source
FROM golang:1.26.0-bookworm AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-strykz ./cmd/nana-boiler

## Lean Image
FROM gcr.io/distroless/base-debian12 AS release-stage

ARG GIT_SHA=unknown
LABEL org.opencontainers.image.revision=$GIT_SHA

WORKDIR /app

COPY --from=build-stage /docker-strykz /docker-strykz

EXPOSE 8080
EXPOSE 6767

ENTRYPOINT ["/docker-strykz"]
""")

# 4. .dockerignore
with open('.dockerignore', 'w') as f:
    f.write(""".env
.git
infra/
**/*_test.go
**/*integration_test.go
""")

# 5. models/orchestrator.go enum
models_orch_path = 'models/orchestrator.go'
with open(models_orch_path, 'r') as f:
    models_orch = f.read()

models_orch = models_orch.replace("""type MatchStatus string

const (
	AwaitingServer MatchStatus = "Awaiting_Server"
)""", """type MatchStatus string

const (
	MatchAwaitingServer MatchStatus = "Awaiting_Server"
	MatchAccepted       MatchStatus = "accepted"
	MatchReady          MatchStatus = "ready"
	MatchFinished       MatchStatus = "finished"
	MatchCancelled      MatchStatus = "cancelled"
)""")
with open(models_orch_path, 'w') as f:
    f.write(models_orch)

# 6. Replace AwaitingServer with MatchAwaitingServer globally
os.system("find . -name '*.go' -type f -exec sed -i 's/models.AwaitingServer/models.MatchAwaitingServer/g' {} +")
os.system("find . -name '*.go' -type f -exec sed -i 's/\"accepted\"/string(models.MatchAccepted)/g' {} +")
os.system("find . -name '*.go' -type f -exec sed -i 's/\"ready\"/string(models.MatchReady)/g' {} +")


# 7. Add TerminateServer to Orchestrator
orch_intf_path = 'internal/services/orchestrator/interface.go'
with open(orch_intf_path, 'r') as f:
    orch_intf = f.read()
if 'TerminateServer' not in orch_intf:
    orch_intf = orch_intf.replace('CreateServer(ctx context.Context, region string) (string, error)',
                                  'CreateServer(ctx context.Context, region string) (string, error)\\n\\tTerminateServer(ctx context.Context, serverID string) error')
    with open(orch_intf_path, 'w') as f:
        f.write(orch_intf)

orch_impl_path = 'internal/services/orchestrator/orchestrator.go'
with open(orch_impl_path, 'r') as f:
    orch_impl = f.read()

if 'TerminateServer(' not in orch_impl:
    orch_impl += """
func (s *Orchestrator) TerminateServer(ctx context.Context, serverID string) error {
	// Terminate the EC2 instance
	_, err := s.ec2client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{serverID},
	})
	if err != nil {
		log.Printf("failed to terminate EC2 instance %s: %v", serverID, err)
		// we might still want to delete it from the DB
	}

	// Delete from our DB
	return s.orchestratorrepo.DeleteServer(ctx, serverID)
}
"""
    with open(orch_impl_path, 'w') as f:
        f.write(orch_impl)

# 8. Call TerminateServer in sidecar.go
sidecar_path = 'internal/grpc/sidecar.go'
with open(sidecar_path, 'r') as f:
    sidecar = f.read()

sidecar = sidecar.replace('''		case *pb.SidecarEvent_ServerStopped:
			log.Println("Here We Would Delete The Server")''', '''		case *pb.SidecarEvent_ServerStopped:
			func() {
				ctx, cancel := context.WithTimeout(stream.Context(), 5*time.Second)
				defer cancel()
				if err := s.orchestrator.TerminateServer(ctx, serverID); err != nil {
					log.Printf("failed to terminate server %s: %v", serverID, err)
				}
			}()''')

with open(sidecar_path, 'w') as f:
    f.write(sidecar)

