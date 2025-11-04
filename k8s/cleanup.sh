#!/bin/bash
kubectl delete -f service.yaml
kubectl delete -f deployment.yaml
kubectl delete -f rbac.yaml
kubectl delete -f serviceaccount.yaml
kubectl delete -f namespace.yaml