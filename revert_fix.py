import os

# fix models/orchestrator.go
models_orch_path = 'models/orchestrator.go'
with open(models_orch_path, 'r') as f:
    models_orch = f.read()

models_orch = models_orch.replace('MatchAccepted       MatchStatus = string(models.MatchAccepted)', 'MatchAccepted       MatchStatus = "accepted"')
models_orch = models_orch.replace('MatchReady          MatchStatus = string(models.MatchReady)', 'MatchReady          MatchStatus = "ready"')
with open(models_orch_path, 'w') as f:
    f.write(models_orch)

# fix internal/services/orchestrator/interface.go
intf_path = 'internal/services/orchestrator/interface.go'
with open(intf_path, 'r') as f:
    intf = f.read()
intf = intf.replace('\\n\\t', '\n\t')
with open(intf_path, 'w') as f:
    f.write(intf)

# fix matchmaking service.go to use proper constants without casting if it's already defined
mm_path = 'internal/services/matchmaking/service.go'
with open(mm_path, 'r') as f:
    mm = f.read()
mm = mm.replace('string(models.MatchReady)', 'models.MatchReady')
mm = mm.replace('string(models.MatchAccepted)', 'models.MatchAccepted')
with open(mm_path, 'w') as f:
    f.write(mm)

