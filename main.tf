
# TODO implement, code is copied from another repo
terraform { #this is we tell terraform we config AWS
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

# TODO creds cleanup
# TODO modularize files 
# TODO is the public subnet safe... is it needed? -- yes because the private subnet routes outbound internet traffic requests to the public subnet
# TODO rename names to have a or b sufixed

provider "aws" { #give user secret access here
  region     = "us-west-2"
  access_key = ""
  secret_key = ""
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
  # ID ???
  tags = {
    Name    = "vpc-mt"
    Purpose = "multi-tenant"
  }
}

# TODO simplify route creation with loop or module
resource "aws_subnet" "public" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.0.0/18"
  availability_zone = "us-west-2d"
  # THe below causes issues if trying to setup nodes on public subnet
  # map_public_ip_on_launch = true # does this do anything, should I delete, not sure its valid for my use-case
  tags = {
    Name                     = "subnet-mt-pub"
    Purpose                  = "multi-tenant"
    "kubernetes.io/role/elb" = 1
    # kubernetes.io/role/elb	
  }
}

resource "aws_subnet" "private" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.64.0/18"
  availability_zone = "us-west-2d"
  tags = {
    Name                              = "subnet-mt-priv"
    Purpose                           = "multi-tenant"
    "kubernetes.io/role/internal-elb" = 1
  }

}

resource "aws_subnet" "publicA" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.128.0/18"
  availability_zone = "us-west-2a"
  # map_public_ip_on_launch = true # does this do anything, should I delete, not sure its valid for my use-case
  tags = {
    Name                     = "subnet-mt-pub"
    Purpose                  = "multi-tenant"
    "kubernetes.io/role/elb" = 1
    # kubernetes.io/role/elb	
  }
}

resource "aws_subnet" "privateA" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.192.0/18"
  availability_zone = "us-west-2a"
  tags = {
    Name                              = "subnet-mt-priv"
    Purpose                           = "multi-tenant"
    "kubernetes.io/role/internal-elb" = 1
  }
}


resource "aws_internet_gateway" "public" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name    = "igw-main"
    Purpose = "multi-tenant"
  }
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name    = "Public Subnet Route Table"
    Purpose = "multi-tenant"
  }
}


resource "aws_route" "igw_route" {
  route_table_id         = aws_route_table.public.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.public.id
}

resource "aws_route_table_association" "public" {
  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.public.id
}

resource "aws_nat_gateway" "public" {
  subnet_id     = aws_subnet.public.id
  allocation_id = aws_eip.public.id
  tags = {
    Name    = "Public NAT"
    Purpose = "multi-tenant"
  }
}

resource "aws_eip" "public" {
}

resource "aws_route_table" "private" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name    = "Private Subnet Route Table"
    Purpose = "multi-tenant"
  }
}


resource "aws_route" "nat_route" {
  route_table_id         = aws_route_table.private.id
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = aws_nat_gateway.public.id
}


resource "aws_route_table_association" "private" {
  subnet_id      = aws_subnet.private.id
  route_table_id = aws_route_table.private.id
}

## AZ a

resource "aws_route_table" "publicA" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name    = "Public Subnet Route Table A"
    Purpose = "multi-tenant"
  }
}


resource "aws_route" "igw_routeA" {
  route_table_id         = aws_route_table.publicA.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.public.id
}

resource "aws_route_table_association" "publicA" {
  subnet_id      = aws_subnet.publicA.id
  route_table_id = aws_route_table.publicA.id
}

resource "aws_nat_gateway" "publicA" {
  subnet_id     = aws_subnet.publicA.id
  allocation_id = aws_eip.publicA.id
  tags = {
    Name    = "Public NAT"
    Purpose = "multi-tenant"
  }
}

resource "aws_eip" "publicA" {
}

resource "aws_route_table" "privateA" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name    = "Private Subnet Route Table"
    Purpose = "multi-tenant"
  }
}


resource "aws_route" "nat_routeA" {
  route_table_id         = aws_route_table.privateA.id
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = aws_nat_gateway.publicA.id
}


resource "aws_route_table_association" "privateA" {
  subnet_id      = aws_subnet.privateA.id
  route_table_id = aws_route_table.privateA.id
}




##### KUBERNETES #####
data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["eks.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "eks" {
  name               = "eks-cluster-example"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json

  tags = {
    Name    = "eks basic role"
    Purpose = "multi-tenant"
  }
}

resource "aws_iam_role_policy_attachment" "eks-AmazonEKSClusterPolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.eks.name
}

# Optionally, enable Security Groups for Pods
# Reference: https://docs.aws.amazon.com/eks/latest/userguide/security-groups-for-pods.html
resource "aws_iam_role_policy_attachment" "eks-AmazonEKSVPCResourceController" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSVPCResourceController"
  role       = aws_iam_role.eks.name
}

resource "aws_eks_cluster" "eks" {
  name     = "eks_multi"
  role_arn = aws_iam_role.eks.arn

  vpc_config {
    subnet_ids = [aws_subnet.public.id, aws_subnet.private.id, aws_subnet.publicA.id, aws_subnet.privateA.id]
  }

  tags = {
    Purpose = "multi-tenant"
  }

  # Ensure that IAM Role permissions are created before and deleted after EKS Cluster handling.
  # Otherwise, EKS will not be able to properly delete EKS managed EC2 infrastructure such as Security Groups.
  depends_on = [
    aws_iam_role_policy_attachment.eks-AmazonEKSClusterPolicy,
    aws_iam_role_policy_attachment.eks-AmazonEKSVPCResourceController,
  ]
}

# TODO find an easier way to install calls to nodegroup module without tofu init being needed

module "base_node_group" {
  source          = "./modules/node_group"
  cluster_name    = aws_eks_cluster.eks.name
  node_group      = {
    name = "base_ng"
  }
}

# module "second_node_group" {
#   source          = "./modules/node_group"
#   cluster_name    = aws_eks_cluster.eks.name
#   node_group      = {
#     name = "second_ng"
#   }
# }