path = 'internal/services/matchmaking/service.go'
with open(path, 'r') as f:
    content = f.read()

content = content.replace('repo.UpdatePlayer(ctx, player, matchID, models.MatchAccepted)', 'repo.UpdatePlayer(ctx, player, matchID, string(models.MatchAccepted))')

with open(path, 'w') as f:
    f.write(content)
