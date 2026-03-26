FROM alpine:3.20

ARG TERRAFORM_VERSION=1.9.8

RUN apk add --no-cache \
    openssh-client \
    ansible \
    curl \
    jq \
    bash \
    git \
    python3 \
    py3-pip \
    unzip \
    dos2unix \
  && pip3 install --no-cache-dir --break-system-packages docker \
  && ansible-galaxy collection install community.docker \
  && curl -fsSL "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip" \
       -o /tmp/terraform.zip \
  && unzip /tmp/terraform.zip -d /usr/local/bin/ \
  && rm /tmp/terraform.zip \
  && terraform version

WORKDIR /workspace

COPY terraform/ ./terraform/
COPY ansible/   ./ansible/
COPY deployer/entrypoint.sh .

RUN dos2unix entrypoint.sh && chmod +x entrypoint.sh

ENTRYPOINT ["./entrypoint.sh"]
