provider "aws" {
  region = "us-east-1"
}

resource "aws_sqs_queue" "orchestrator_queue_fifo" {
  name                        = "orchestrator_queue.fifo"
  fifo_queue                  = true
  content_based_deduplication = true
}
