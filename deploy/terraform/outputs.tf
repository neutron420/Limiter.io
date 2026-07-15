output "public_ip" {
  description = "Elastic IP of the host — point your DNS here"
  value       = aws_eip.limiter.public_ip
}

output "ssh_command" {
  description = "SSH into the instance"
  value       = "ssh -i ~/.ssh/limiter_deploy ubuntu@${aws_eip.limiter.public_ip}"
}

output "landing_url" {
  value = "http://${aws_eip.limiter.public_ip}"
}

output "api_url" {
  value = "http://${aws_eip.limiter.public_ip}/api/v1"
}
