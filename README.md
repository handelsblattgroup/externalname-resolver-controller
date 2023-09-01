ExternalName Resolver Controller
====

[![1. Stable Build, Test and Push](https://github.com/handelsblattgroup/externalname-resolver-controller/actions/workflows/1_stable.yml/badge.svg)](https://github.com/handelsblattgroup/externalname-resolver-controller/actions/workflows/1_stable.yml)

Kubernetes controller to manage ExternalName type Services resolving their DNS entry, generating and keeping in sync specific Endpoints.


## Purpose

Nowadays you can't use ExternalName Services inside Ingress rules because they are missing Endpoint resources.

The workaround is to create a ClusterIP type Service with no selector and to manually add an Endpoint with the same resource name and a resolved IP from the ExternalName DNS entry.

This controller automates this process by reacting to annotations or forcing the conversion of ExternalName type Service to this workaround.


