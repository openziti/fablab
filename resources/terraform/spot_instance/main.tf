variable "access_key" {}
variable "secret_key" {}
variable "amis" {}
variable "environment_tag" { default = "" }
variable "instance_type" {}
variable "key_name" {}
variable "key_path" {}
variable "region" {}
variable "security_group_id" {}
variable "ssh_user" { default = "ubuntu" }
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

  provisioner "remote-exec" {
    connection {
      host        = self.public_ip
      type        = "ssh"
      agent       = false
      user        = var.ssh_user
      private_key = file(var.key_path)
    }

    inline = [
      "sudo chmod 777 /etc/sysctl.d",
    ]
  }

  provisioner "file" {
    connection {
      host        = self.public_ip
      type        = "ssh"
      agent       = false
      user        = var.ssh_user
      private_key = file(var.key_path)
    }

    source        = "etc/apt/apt.conf.d/99remote-not-fancy"
    destination   = "/home/ubuntu/99remote-not-fancy"
  }

  provisioner "file" {
    connection {
      host        = self.public_ip
      type        = "ssh"
      agent       = false
      user        = var.ssh_user
      private_key = file(var.key_path)
    }

    source        = "etc/sysctl.d/51-network-tuning.conf"
    destination   = "/etc/sysctl.d/51-network-tuning.conf"
  }

  provisioner "remote-exec" {
    connection {
      host        = self.public_ip
      type        = "ssh"
      agent       = false
      user        = var.ssh_user
      private_key = file(var.key_path)
    }

    inline = [
      "sudo mv /home/ubuntu/99remote-not-fancy /etc/apt/apt.conf.d/",
      "sudo chmod 755 /etc/sysctl.d",
      "sudo apt update",
      "sudo apt upgrade -y",
      "sudo apt install -y iperf3 tcpdump sysstat",
      "sudo bash -c \"echo 'ubuntu soft nofile 40960' >> /etc/security/limits.conf\"",
      "sudo shutdown -r +1"
    ]
  }
}