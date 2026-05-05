#!/usr/bin/env bash
set -euo pipefail

# 1. Create a new k3d cluster (exposes 80 → 30949, 443 → 30950; 2 worker agents)
k3d cluster create hlf-local \
  -p "80:30949@agent:0" \
  -p "443:30950@agent:0" \
  --agents 2 \
  --wait
# :contentReference[oaicite:0]{index=0}

#update the server port
kubectl config set-cluster k3d-hlf-local \
  --server=https://127.0.0.1:543637

# 2. Install the Fabric operator via Helm
helm repo add kfs https://kfsoftware.github.io/hlf-helm-charts --force-update
helm repo update
helm install hlf-operator --version=1.11.1 kfs/hlf-operator
# :contentReference[oaicite:1]{index=1}
export PATH="${KREW_ROOT:-$HOME/.krew}/bin:$PATH"
# 3. Install the 'kubectl hlf' plugin (requires krew)
kubectl krew install hlf
# :contentReference[oaicite:2]{index=2}

# 4. Install Istio and enable the ingress gateway
ISTIO_VERSION=1.23.3
curl -L https://istio.io/downloadIstio | ISTIO_VERSION=${ISTIO_VERSION} sh -
kubectl create namespace istio-system
export ISTIO_PATH="$PWD/istio-${ISTIO_VERSION}/bin"
export PATH="$PATH:$ISTIO_PATH"
istioctl operator init

cat <<EOF | kubectl apply -f -
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: istio-gateway
  namespace: istio-system
spec:
  addonComponents:
    grafana: { enabled: false }
    kiali:   { enabled: false }
    prometheus: { enabled: false }
    tracing:    { enabled: false }
  components:
    ingressGateways:
      - name: istio-ingressgateway
        enabled: true
        k8s:
          hpaSpec: { minReplicas: 1 }
          resources:
            requests: { cpu: 100m, memory: 128Mi }
            limits:   { cpu: 500m, memory: 512Mi }
          service:
            type: NodePort
            ports:
              - name: http
                port: 80
                targetPort: 8080
                nodePort: 30949
              - name: https
                port: 443
                targetPort: 8443
                nodePort: 30950
    pilot:
      enabled: true
      k8s:
        hpaSpec: { minReplicas: 1 }
        resources:
          requests: { cpu: 100m, memory: 128Mi }
          limits:   { cpu: 300m, memory: 512Mi }
  meshConfig:
    accessLogFile: /dev/stdout
    enableTracing: false
    outboundTrafficPolicy: { mode: ALLOW_ANY }
  profile: default
EOF
# :contentReference[oaicite:3]{index=3}

# 5. Set Fabric image variables and storage class for k3d
export PEER_IMAGE=hyperledger/fabric-peer
export PEER_VERSION=3.0.0
export ORDERER_IMAGE=hyperledger/fabric-orderer
export ORDERER_VERSION=3.0.0
export CA_IMAGE=hyperledger/fabric-ca
export CA_VERSION=1.5.13
export SC_NAME=local-path
# :contentReference[oaicite:4]{index=4}

# 6. Configure internal DNS so *.localho.st → Istio ingress
cat <<EOF | kubectl apply -f -
kind: ConfigMap
apiVersion: v1
metadata:
  name: coredns
  namespace: kube-system
data:
  Corefile: |
    .:53 {
        errors
        health { lameduck 5s }
        rewrite name regex (.*)\.localho\.st istio-ingressgateway.istio-system.svc.cluster.local
        hosts { fallthrough }
        kubernetes cluster.local in-addr.arpa ip6.arpa {
          pods insecure
          fallthrough in-addr.arpa ip6.arpa
          ttl 30
        }
        forward . /etc/resolv.conf { max_concurrent 1000 }
        cache 30; loop; reload; loadbalance
    }
EOF
# :contentReference[oaicite:5]{index=5}

# 7. Deploy Org1’s CA, register peer user, and spin up peer0
kubectl hlf ca create \
  --image=$CA_IMAGE --version=$CA_VERSION \
  --storage-class=$SC_NAME --capacity=1Gi \
  --name=org1-ca \
  --enroll-id=enroll --enroll-pw=enrollpw \
  --hosts=org1-ca.localho.st \
  --istio-port=443

kubectl wait --timeout=180s --for=condition=Running fabriccas.hlf.kungfusoftware.es --all

# sanity check:
curl -k https://org1-ca.localho.st:443/cainfo

kubectl hlf ca register \
  --name=org1-ca \
  --user=peer --secret=peerpw \
  --type=peer \
  --enroll-id=enroll --enroll-secret=enrollpw \
  --mspid=Org1MSP

kubectl hlf peer create \
  --statedb=leveldb \
  --image=$PEER_IMAGE --version=$PEER_VERSION \
  --storage-class=$SC_NAME \
  --enroll-id=peer --mspid=Org1MSP --enroll-pw=peerpw \
  --capacity=5Gi \
  --name=org1-peer0 \
  --ca-name=org1-ca.default \
  --hosts=peer0-org1.localho.st \
  --istio-port=443

kubectl wait --timeout=180s --for=condition=Running fabricpeers.hlf.kungfusoftware.es --all
# :contentReference[oaicite:6]{index=6}

# 8. Deploy Orderer Org: CA → register → 4 orderer nodes
kubectl hlf ca create \
  --image=$CA_IMAGE --version=$CA_VERSION \
  --storage-class=$SC_NAME --capacity=1Gi \
  --name=ord-ca \
  --enroll-id=enroll --enroll-pw=enrollpw \
  --hosts=ord-ca.localho.st \
  --istio-port=443

kubectl wait --timeout=180s --for=condition=Running fabriccas.hlf.kungfusoftware.es --all

curl -k https://ord-ca.localho.st:443/cainfo

kubectl hlf ca register \
  --name=ord-ca \
  --user=orderer --secret=ordererpw \
  --type=orderer \
  --enroll-id=enroll --enroll-secret=enrollpw \
  --mspid=OrdererMSP \
  --ca-url="https://ord-ca.localho.st:443"

for i in {1..4}; do
  kubectl hlf ordnode create \
    --image=$ORDERER_IMAGE --version=$ORDERER_VERSION \
    --storage-class=$SC_NAME \
    --enroll-id=orderer --mspid=OrdererMSP --enroll-pw=ordererpw \
    --capacity=2Gi \
    --name=ord-node${i} \
    --ca-name=ord-ca.default \
    --hosts=orderer$((i-1))-ord.localho.st \
    --admin-hosts=admin-orderer$((i-1))-ord.localho.st \
    --istio-port=443
done

kubectl wait --timeout=180s --for=condition=Running fabricorderernodes.hlf.kungfusoftware.es --all
# :contentReference[oaicite:7]{index=7}

# 9. Create a channel (‘demo’) and join peer
# 9a. Register & enroll admin identities for OrdererMSP
kubectl hlf ca register --name=ord-ca --user=admin --secret=adminpw --type=admin --enroll-id=enroll --enroll-secret=enrollpw --mspid=OrdererMSP
kubectl hlf ca enroll  --name=ord-ca --namespace=default --user=admin --secret=adminpw --mspid=OrdererMSP --ca-name=tlsca --output orderermsp.yaml
kubectl hlf ca enroll  --name=ord-ca --namespace=default --user=admin --secret=adminpw --mspid=OrdererMSP --ca-name=ca    --output orderermspsign.yaml

# 9b. Register & enroll admin identities for Org1MSP
kubectl hlf ca register --name=org1-ca --user=admin --secret=adminpw --type=admin --enroll-id=enroll --enroll-secret=enrollpw --mspid=Org1MSP
kubectl hlf ca enroll  --name=org1-ca --namespace=default --user=admin --secret=adminpw --mspid=Org1MSP --ca-name=tlsca --output org1msp-tlsca.yaml
kubectl hlf ca enroll  --name=org1-ca --namespace=default --user=admin --secret=adminpw --mspid=Org1MSP --ca-name=ca    --output org1msp.yaml

kubectl hlf identity create --name org1-admin --namespace default --ca-name org1-ca --ca-namespace default --ca ca --mspid Org1MSP --enroll-id admin --enroll-secret adminpw

kubectl create secret generic wallet --namespace=default \
  --from-file=org1msp.yaml=$PWD/org1msp.yaml \
  --from-file=orderermsp.yaml=$PWD/orderermsp.yaml \
  --from-file=orderermspsign.yaml=$PWD/orderermspsign.yaml

# 9c. Generate channel config and create it
export PEER_ORG_SIGN_CERT=$(kubectl get fabriccas org1-ca -o=jsonpath='{.status.ca_cert}')
export PEER_ORG_TLS_CERT=$(kubectl get fabriccas org1-ca -o=jsonpath='{.status.tlsca_cert}')
export IDENT_8=$(printf "%8s")
export ORDERER_TLS_CERT=$(kubectl get fabriccas ord-ca -o=jsonpath='{.status.tlsca_cert}' | sed -e "s/^/${IDENT_8}/")
export ORDERER0_TLS_CERT=$(kubectl get fabricorderernodes ord-node1 -o=jsonpath='{.status.tlsCert}' | sed -e "s/^/${IDENT_8}/")
export ORDERER1_TLS_CERT=$(kubectl get fabricorderernodes ord-node2 -o=jsonpath='{.status.tlsCert}' | sed -e "s/^/${IDENT_8}/")
export ORDERER2_TLS_CERT=$(kubectl get fabricorderernodes ord-node3 -o=jsonpath='{.status.tlsCert}' | sed -e "s/^/${IDENT_8}/")
export ORDERER3_TLS_CERT=$(kubectl get fabricorderernodes ord-node4 -o=jsonpath='{.status.tlsCert}' | sed -e "s/^/${IDENT_8}/")

kubectl apply -f - <<EOF
apiVersion: hlf.kungfusoftware.es/v1alpha1
kind: FabricMainChannel
metadata:
  name: demo
spec:
  name: demo
  adminOrdererOrganizations:
    - mspID: OrdererMSP
  adminPeerOrganizations:
    - mspID: Org1MSP
  channelConfig:
    application:
      capabilities: [V2_0,V2_5]
    capabilities: [V2_0]
    orderer:
      batchSize:
        absoluteMaxBytes: 1048576
        maxMessageCount: 10
        preferredMaxBytes: 524288
      batchTimeout: 2s
      capabilities: [V2_0]
      etcdRaft:
        options:
          electionTick: 10
          heartbeatTick: 1
          maxInflightBlocks: 5
          snapshotIntervalSize: 16777216
          tickInterval: 500ms
      ordererType: etcdraft
  peerOrganizations:
    - mspID: Org1MSP
      caName: org1-ca
      caNamespace: default
      identities:
        Org1MSP:
          secretKey: org1msp.yaml
          secretName: wallet
          secretNamespace: default
  ordererOrganizations:
    - mspID: OrdererMSP
      caName: ord-ca
      caNamespace: default
      externalOrderersToJoin:
        - host: ord-node1.default port:7053
        - host: ord-node2.default port:7053
        - host: ord-node3.default port:7053
        - host: ord-node4.default port:7053
  orderers:
    - host: orderer0-ord.localho.st port:443 tlsCert: |-\${ORDERER0_TLS_CERT}
    - host: orderer1-ord.localho.st port:443 tlsCert: |-\${ORDERER1_TLS_CERT}
    - host: orderer2-ord.localho.st port:443 tlsCert: |-\${ORDERER2_TLS_CERT}
    - host: orderer3-ord.localho.st port:443 tlsCert: |-\${ORDERER3_TLS_CERT}
EOF

# 9d. Join Org1 peer to 'demo'
export IDENT_8=$(printf "%8s")
export ORDERER0_TLS_CERT=$(kubectl get fabricorderernodes ord-node1 -o=jsonpath='{.status.tlsCert}' | sed -e "s/^/${IDENT_8}/")

kubectl apply -f - <<EOF
apiVersion: hlf.kungfusoftware.es/v1alpha1
kind: FabricFollowerChannel
metadata:
  name: demo-org1msp
spec:
  mspId: Org1MSP
  name: demo
  anchorPeers:
    - host: peer0-org1.localho.st port:443
  hlfIdentity:
    secretKey: org1msp.yaml
    secretName: wallet
    secretNamespace: default
  orderers:
    - certificate: |-\${ORDERER0_TLS_CERT}
      url: grpcs://ord-node1.default:7050
  peersToJoin:
    - name: org1-peer0 namespace: default
EOF
# :contentReference[oaicite:8]{index=8}

echo "✅ Hyperledger Fabric network up and running on k3d (cluster 'hlf-local')!"
