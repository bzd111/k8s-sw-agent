openssl genrsa -out ca.key 2048
openssl req -x509 -new -nodes -key ca.key -subj "/CN=pod-admission-webhook.kube-system.svc" -days 10000 -out ca.crt
openssl genrsa -out tls.key 2048
openssl req -new -key tls.key -out tls.csr -config csr.conf
openssl x509 -req -in tls.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out tls.crt -days 10000 -extensions v3_ext -extfile csr.conf

kubectl create secret tls pod-admission-webhook --dry-run=client --cert=tls.crt --key=tls.key --namespace kube-system -o yaml > secret.yaml

cabundle=openssl base64 -A <"ca.crt"
sed -i "s/cabundle/${cabundle}/g" ./mutatingwebhook.yaml 
