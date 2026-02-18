## Unreleased

- 2026-02-18: Fix: Guard against nil pointer when comparing Deployment.Spec.Replicas in internal/controller/helloapp_controller.go to avoid a potential panic when an existing Deployment has a nil `Replicas` pointer.
