output "beanstalk_url" {
  value = aws_elastic_beanstalk_environment.backend_env.endpoint_url
}

output "rds_endpoint" {
  value = aws_db_instance.postgres.endpoint
}

output "redis_private_ip" {
  value = aws_instance.redis.private_ip
}
