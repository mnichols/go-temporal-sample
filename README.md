# go temporal sample

This simple sample demonstrates how to execute a workflow that, upon encountering a failure, notifies a user so that they might correct the erroneous input.

## To Simulate

Start local dev server as: `temporal server start-dev`

Then in another terminal start the worker: `go run cmd/backend/main.go`

Finally issue these commands to spoof a failure:


```bash

# start the workflow with a bad SubscriptionID of "GarbageSubscriptionID"
temporal workflow start --type OnboardApplication \
    --workflow-id foo \
    --input-file onboard_application.json
    
# observe that:
# 1. the "SetupJFrog" activity has failed due to having this SubscriptionID
# 2. the workflow detects this BadSubscriptionIDErr so RequestsCorrection  
    
# send a signal to the workflow to correct the bad SubscriptionID
temporal workflow signal --name correctSubscriptionID \
    --input '{"subscriptionId": "123"}' 
    --workflow-id foo
    
# observe that the workflow recalls the SetupJFrog activity and exits
```
