output "sqs_queue_url" {
  value = aws_sqs_queue.orchestrator_queue_fifo.url
}

output "sqs_queue_arn" {
  value = aws_sqs_queue.orchestrator_queue_fifo.arn
}

output "sqs_queue_name" {
  value = aws_sqs_queue.orchestrator_queue_fifo.name
}
