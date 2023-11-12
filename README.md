# virtual-friends-brain

## Deployments

### Build images
```sh
# main container
gcloud builds --project ysong-chat submit --tag gcr.io/ysong-chat/brain-go:$(git rev-parse --short HEAD)-$(openssl rand -hex 4) .

# python sidecar
cd flask-py
gcloud builds --project ysong-chat submit --tag gcr.io/ysong-chat/flask-app:$(git rev-parse --short HEAD)-$(openssl rand -hex 4) .
```

### Deploy
The two `gcloud builds` commands will output the image paths:
```sh
38fc5f0e-855e-4dad-8218-630c7bf6bdee  2023-10-25T03:39:15+00:00  3M17S     gs://ysong-chat_cloudbuild/source/1698205154.301932-b9b2af457c1041399df99f7afd8143f9.tgz  gcr.io/ysong-chat/brain-go:bd130ef-56167d45  SUCCESS

256f0808-9fb6-48e2-929c-898c897369dd  2023-11-12T00:43:15+00:00  7M53S     gs://ysong-chat_cloudbuild/source/1699749793.760277-ca9ceb32d5df4ab29ef8eb15923830ad.tgz  gcr.io/ysong-chat/flask-app:2581cfb-6f8100ca  SUCCESS
```
Copy and paste the two image paths *gcr.io/ysong-chat/...* to the ./k8s/deployments.yaml's corresponding places and run
```sh
kubectl apply -f k8s/deployment.yaml
```
And delete the existing pod.
```sh
kubectl get pods
kubectl delete pods <virtual-friends-gke- current running pod>
```
