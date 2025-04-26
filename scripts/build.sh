#!/bin/bash
set -e

# Step 1: Create a local kubernetes cluster if you don't have one
kind create cluster --name fabric-local

# Step 2: Clone the Fabric Operator repository
git clone https://github.com/hyperledger-labs/fabric-operator.git
cd fabric-operator

# Step 3: Build the operator
# This builds the operator and installs CRDs
make install

# Step 4: Build and push Docker image to your local Kind cluster
# Build the operator image
make docker-build IMG=fabric-operator:latest

# Load the image into your Kind cluster
kind load docker-image fabric-operator:latest --name fabric-local

# Step 5: Deploy the operator
# Create a namespace for the operator
kubectl create ns fabric-operator

# Deploy the operator
make deploy IMG=fabric-operator:latest NAMESPACE=fabric-operator

# Step 6: Verify installation
kubectl get pods -n fabric-operator
kubectl get crds | grep ibp

# Step 7: Only after confirming operator is running, deploy a test CA and peer
cat <<'EOF' | kubectl apply -f -
apiVersion: ibp.com/v1
kind: IBPCA
metadata:
  name: org1-ca
  namespace: fabric-operator
spec:
  license: accept
  replicas: 1
---
apiVersion: ibp.com/v1
kind: IBPPeer
metadata:
  name: org1-peer0
  namespace: fabric-operator
spec:
  license: accept
  images:
    caInitImage: ghcr.io/hyperledger/fabric-ca:1.5
    peerInitImage: ghcr.io/hyperledger/fabric-peer:2.5
  storage:
    peer:
      class: standard
      size: 10Gi
EOF