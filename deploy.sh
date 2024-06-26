gcloud functions deploy fuftyfu-api \
    --gen2 \
    --trigger-http \
    --runtime=go121 \
    --entry-point=main \
    --region=europe-central2 \
    --source=. \
    --allow-unauthenticated \
    --project zinovik-project \
    --set-secrets=TOKEN=fuftyfy-api-app-token:latest
