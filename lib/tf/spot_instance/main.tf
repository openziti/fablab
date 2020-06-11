variable "access_key" {}
variable "secret_key" {}
variable "amis" {}
variable "environment_tag" { default = "" }
variable "instance_type" {}
variable "key_name" {}
variable "key_path" {}
variable "region" {}
variable "security_group_id" {}
variable "ssh_user" { default = "fedora" }
variable "subnet_id" {}
variable "spot_price" {}
variable "spot_type" {}

output "public_ip" { value = aws_spot_instance_request.fablab.public_ip }
output "private_ip" { value = aws_spot_instance_request.fablab.private_ip }

provider "aws" {
  access_key = var.access_key
  secret_key = var.secret_key
  region     = var.region
}

resource "aws_spot_instance_request" "fablab" {
  ami                         = lookup(var.amis, var.region)
  instance_type               = var.instance_type
  key_name                    = var.key_name
  vpc_security_group_ids      = [var.security_group_id]
  subnet_id                   = var.subnet_id
  associate_public_ip_address = true

  // spot instance-specific args
  spot_price                  = var.spot_price
  wait_for_fulfillment        = true
  spot_type                   = var.spot_type

  /** 
    The following directive doesn't work for spot instances, so we 'aws ec2 create-tags ...' in the 'inline' section below

  tags = {
    Name = var.environment_tag
  }

  */

  provisioner "remote-exec" {
    connection {
      host        = self.public_ip
      type        = "ssh"
      agent       = false
      user        = var.ssh_user
      private_key = file(var.key_path)
    }

    inline = [
      "sudo dnf update -y",
      "sudo dnf install -y iperf3 tcpdump sysstat",
      
      // Install, then use, the AWS CLI to apply a Name tag to this instance
      "pip3 install awscli --upgrade --user",
      "mkdir -p /home/fedora/.aws",
      "sudo printf '[default]\naws_access_key_id = ${var.access_key}\naws_secret_access_key = ${var.secret_key}\n' | sudo tee /home/fedora/.aws/credentials",
      "aws ec2 create-tags --region ${var.region} --resources ${aws_spot_instance_request.fablab.spot_instance_id} --tags Key=Name,Value=${var.environment_tag}",

      "sudo bash -c \"echo 'fedora soft nofile 40960' >> /etc/security/limits.conf\"",
      "sudo shutdown -r +1"
    ]
  }


}