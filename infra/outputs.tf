output "beanstalk_url" {
  value = aws_elastic_beanstalk_environment.backend_env.endpoint_url
}

output "rds_endpoint" {
  value = aws_db_instance.postgres.endpoint
}

output "game_server_ip" {
  value = aws_instance.game_server.public_ip
}

output "redis_private_ip" {
  value = aws_instance.redis.private_ip
}