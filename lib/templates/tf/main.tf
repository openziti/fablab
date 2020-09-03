variable "environment_tag"    { default = "{{ .Model.MustVariable "environment" }}" }
variable "aws_access_key"     { default = "{{ .Model.MustVariable "credentials" "aws" "access_key" }}" }
variable "aws_secret_key"     { default = "{{ .Model.MustVariable "credentials" "aws" "secret_key" }}" }
variable "aws_key_name"       { default = "{{ .Model.MustVariable "credentials" "aws" "ssh_key_name" }}" }
variable "aws_key_path"       { default = "{{ .Model.MustVariable "credentials" "ssh" "key_path" }}" }

variable "vpc_cidr"           { default = "10.0.0.0/16" }
variable "public_cidr"        { default = "10.0.0.0/24" }

/*
 * Fedora 30 Standard Base Cloud HVM
*/
variable "amis" {
  default = {
    us-east-1      = "ami-00bbc6858140f19ed"
    us-east-2      = "ami-00b43acad6bfbc73a"
    us-west-2      = "ami-0faa715e16d9b8cfc"
    us-west-1      = "ami-0b03257899cb1d721"
    eu-west-1      = "ami-050e739238754544e"
    eu-central-1   = "ami-040a3cdf71cb042c0"
    eu-west-2      = "ami-06e02c487847a397f"
    eu-west-3      = "ami-04abc80987c29d6fb"
    ap-southeast-1 = "ami-072318105f445e4e6"
    ap-northeast-1 = "ami-0994d322d1354126a"
    ap-southeast-2 = "ami-067e0a8a8c3329ceb"
    sa-east-1      = "ami-028ce5f4a8e41f793"
    ap-northeast-2 = "ami-083f78394597eac33"
    ap-south-1     = "ami-0e4e37dcffc88ad29"
    ca-central-1   = "ami-02ffb9dd4d8c6e863"
  }
}

{{ range $regionId, $region := .Model.Regions }}
module "{{ $regionId }}_region" {
  source          = "{{ $.TerraformLib }}/vpc"
  access_key      = var.aws_access_key
  secret_key      = var.aws_secret_key
  region          = "{{ $region.Region }}"
  vpc_cidr        = var.vpc_cidr
  public_cidr     = var.public_cidr
  az              = "{{ $region.Site }}"
  environment_tag = var.environment_tag
}
{{ range $hostId, $host := $region.Hosts }}
module "{{ $regionId }}_host_{{ $hostId }}" {
  source            = "{{ $.TerraformLib }}/{{ instanceTemplate $host }}_instance"
  access_key        = var.aws_access_key
  secret_key        = var.aws_secret_key
  amis              = var.amis
  environment_tag   = var.environment_tag
  instance_type     = "{{ $host.InstanceType }}"
  key_name          = var.aws_key_name
  key_path          = var.aws_key_path
  region            = "{{ $region.Region }}"
  security_group_id = module.{{ $regionId }}_region.security_group_id
  subnet_id         = module.{{ $regionId }}_region.subnet_id
  spot_price        = "{{ $host.SpotPrice }}"
  spot_type         = "{{ $host.SpotType }}"
}

output "{{ $regionId }}_host_{{ $hostId }}_public_ip" { value = module.{{ $regionId }}_host_{{ $hostId }}.public_ip }
output "{{ $regionId }}_host_{{ $hostId }}_private_ip" { value = module.{{ $regionId }}_host_{{ $hostId }}.private_ip }
{{ end }}
{{ end }}
