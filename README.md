## google cloud setup

### create service accounts

```bash
gcloud iam service-accounts create github-actions
gcloud iam service-accounts create fuftyfu
```

### add roles (`Service Account User` and `Cloud Functions Admin`) to the service account you want to use to deploy the function

```
gcloud projects add-iam-policy-binding zinovik-project --member="serviceAccount:github-actions@zinovik-project.iam.gserviceaccount.com" --role="roles/cloudfunctions.admin"

gcloud projects add-iam-policy-binding zinovik-project --member="serviceAccount:github-actions@zinovik-project.iam.gserviceaccount.com" --role="roles/iam.serviceAccountUser"
```

### creating keys for service account for github-actions `GOOGLE_CLOUD_SERVICE_ACCOUNT_KEY_FILE`

```bash
gcloud iam service-accounts keys create key-file.json --iam-account=github-actions@zinovik-project.iam.gserviceaccount.com
cat key-file.json | base64
```

### add access to secrets and bucket

```
gcloud projects add-iam-policy-binding zinovik-project --member="serviceAccount:fuftyfu@zinovik-project.iam.gserviceaccount.com" --role="roles/secretmanager.secretAccessor"

gcloud storage buckets add-iam-policy-binding gs://hedgehogs --member="serviceAccount:fuftyfu@zinovik-project.iam.gserviceaccount.com" --role="roles/storage.admin"
```

### add secrets

```
printf "TOKEN" | gcloud secrets create fuftyfy-api-app-token --locations=europe-central2 --replication-policy="user-managed" --data-file=-
```
