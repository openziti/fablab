variable "environment_tag"    { default = "{{ .Model.MustVariable "environment" }}" }
variable "aws_access_key"     { default = "{{ .Model.MustVariable "credentials.aws.access_key" }}" }
variable "aws_secret_key"     { default = "{{ .Model.MustVariable "credentials.aws.secret_key" }}" }
variable "aws_key_name"       { default = "{{ .Model.MustVariable "credentials.aws.ssh_key_name" }}" }
variable "aws_key_path"       { default = "{{ .Model.MustVariable "credentials.ssh.key_path" }}" }

variable "vpc_cidr"           { default = "10.0.0.0/16" }
variable "public_cidr"        { default = "10.0.0.0/24" }

/*
 * Ubuntu 20.04 LTS (Focal Fossa)
*/
variable "amis" {
  default = {
	af-south-1		= "ami-08a4b40f2fe1e4b35"
	ap-east-1		= "ami-0b215afe809665ae5"
	ap-northeast-1	= "ami-0df99b3a8349462c6"
	ap-northeast-2	= "ami-04876f29fd3a5e8ba"
	ap-northeast-3	= "ami-0001d1dd884af8872"
	ap-south-1		= "ami-0c1a7f89451184c8b"
	ap-southeast-1	= "ami-0d058fe428540cd89"
	ap-southeast-2	= "ami-0567f647e75c7bc05"
	ca-central-1	= "ami-0a6aabb67a28328df"
	cn-north-1		= "ami-00e7797a8e3c1f7f6"
	cn-northwest-1	= "ami-0beff0eca7fd2e2c5"
	eu-central-1	= "ami-05f7491af5eef733a"
	eu-north-1		= "ami-0ff338189efb7ed37"
	eu-south-1		= "ami-018f430e4f5375e69"
	eu-west-1		= "ami-0a8e758f5e873d1c1"
	eu-west-2		= "ami-0194c3e07668a7e36"
	eu-west-3		= "ami-0f7cd40eac2214b37"
	me-south-1		= "ami-0eddb8cfbd6a5f657"
	sa-east-1		= "ami-054a31f1b3bf90920"
	us-east-1		= "ami-0464ed28453549213"
	us-east-2		= "ami-00399ec92321828f5"
	us-gov-east-1	= "ami-0dec4096f1af85e9b"
	us-gov-west-1	= "ami-0c39aacd1cc8a1ccf"
	us-west-1		= "ami-0d382e80be7ffdae5"
	us-west-2		= "ami-0329ee6c35395b766"
  }
}

/*
  Ubuntu 21.04
*/
/*
variable "amis" {
  default = {
    af-south-1 = "ami-073a4a2bac101e6c8"
    ap-east-1 = "ami-092ecf731cc60eb49"
    ap-northeast-1 = "ami-06a79c94d0aa293b0"
    ap-northeast-2 = "ami-0ecb7883725e063f5"
    ap-northeast-3 = "ami-0169e0d529ec00f8a"
    ap-south-1 = "ami-0f733fbd96037f20b"
    ap-southeast-1 = "ami-0c0c3cdcce6197c04"
    ap-southeast-2 = "ami-0b106dd9a74650502"
    ca-central-1 = "ami-01d9eb3ae060af8fc"
    eu-central-1 = "ami-056dae67562399d1b"
    eu-north-1 = "ami-0f5664af7f369f6da"
    eu-south-1 = "ami-0436b309681d70382"
    eu-west-1 = "ami-00622de605b25a7d6"
    eu-west-2 = "ami-07cc113fcdecaf20f"
    eu-west-3 = "ami-0703df43147ac855c"
    me-south-1 = "ami-0acbc1b0f7d4cbce7"
    sa-east-1 = "ami-037d0027548e093f2"
    us-east-1 = "ami-0f387434d1dfc4cc2"
    us-east-2 = "ami-01e80d9e7b6daabcf"
    us-west-1 = "ami-027e3197838dcd040"
    us-west-2 = "ami-035649ffeb04ce758"
    cn-north-1= "ami-0e4526fd9c19bed04"
    cn-northwest-1 = "ami-017a6d3f462daf7a4"
    us-gov-east-1 = "ami-014371de1c2e154c6"
    us-gov-west-1 = "ami-03d3bdb061e68a7aa"
  }
}
*/

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
