variable "aws_region" {
  description = "AWS region to deploy into"
  type        = string
  default     = "ap-south-1"
}

variable "project_name" {
  description = "Name prefix for all resources"
  type        = string
  default     = "limiter-io"
}

variable "instance_type" {
  description = "EC2 instance type. t3.medium+ recommended (Kafka + Postgres + Redis + API + Next.js)."
  type        = string
  default     = "t3.micro"
}

variable "root_volume_gb" {
  description = "Root EBS volume size in GB"
  type        = number
  default     = 30
}

variable "ssh_public_key" {
  description = "Public key used for SSH + GitHub Actions deploys (contents of e.g. ~/.ssh/limiter_deploy.pub)"
  type        = string
}

variable "ssh_allowed_cidr" {
  description = "CIDR allowed to SSH. Set to YOUR-IP/32 for safety; GitHub Actions needs 0.0.0.0/0 unless you use a fixed runner."
  type        = string
  default     = "0.0.0.0/0"
}
