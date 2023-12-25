gcloud functions deploy fuftyfu-api \
    --gen2 \
    --trigger-http \
    --runtime=go121 \
    --entry-point=main \
    --region=us-central1 \
    --source=. \
    --allow-unauthenticated \
    --project zinovik-project \
     --env-vars-file env.yaml
