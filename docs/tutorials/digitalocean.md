# Setting up ExternalDNS for Services on DigitalOcean

This tutorial describes how to setup ExternalDNS for usage within a Kubernetes cluster using DigitalOcean DNS.

Make sure to use **>=0.4.0** version of ExternalDNS for this tutorial.

## Creating a DigitalOcean DNS zone

If you want to learn about how to use DigitalOcean's DNS service read the following tutorial series:

[An Introduction to Managing DNS](https://www.digitalocean.com/community/tutorial_series/an-introduction-to-managing-dns), and specifically [How To Set Up a Host Name with DigitalOcean DNS](https://www.digitalocean.com/community/tutorials/how-to-set-up-a-host-name-with-digitalocean)

Create a new DNS zone where you want to create your records in. Let's use `example.com` as an example here.

## Creating DigitalOcean Credentials

Generate a new personal token by going to [the API settings](https://cloud.digitalocean.com/settings/api/tokens) or follow [How To Use the DigitalOcean API v2](https://www.digitalocean.com/community/tutorials/how-to-use-the-digitalocean-api-v2) if you need more information. Give the token a name and choose read and write access. The token needs to be passed to ExternalDNS so make a note of it for later use.

The environment variable `DO_TOKEN` will be needed to run ExternalDNS with DigitalOcean.

## Deploy ExternalDNS

Create a deployment file called `externaldns.yaml` with the following contents:

```yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: external-dns
spec:
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: external-dns
    spec:
      containers:
      - name: external-dns
        image: registry.opensource.zalan.do/teapot/external-dns:v0.4.0
        args:
        - --source=service # ingress is also possible
        - --domain-filter=example.com # (optional) limit to only example.com domains; change to match the zone created above.
        - --provider=digitalocean
        env:
        - name: DO_TOKEN
          value: "YOUR_DIGITALOCEAN_API_KEY"
```

Create the deployment for ExternalDNS:

```console
$ kubectl create -f externaldns.yaml
```

## Deploying an Nginx Service

Create a service file called 'nginx.yaml' with the following contents:

```yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx
        name: nginx
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: nginx
  annotations:
    external-dns.alpha.kubernetes.io/hostname: my-app.example.com
spec:
  selector:
    app: nginx
  type: LoadBalancer
  ports:
    - protocol: TCP
      port: 80
      targetPort: 80
```

Note the annotation on the service; use the same hostname as the DigitalOcean DNS zone created above.

ExternalDNS uses this annotation to determine what services should be registered with DNS. Removing the annotation will cause ExternalDNS to remove the corresponding DNS records.

Create the deployment and service:

```console
$ kubectl create -f nginx.yaml
```

Depending where you run your service it can take a little while for your cloud provider to create an external IP for the service.

Once the service has an external IP assigned, ExternalDNS will notice the new service IP address and synchronize the DigitalOcean DNS records.

## Verifying DigitalOcean DNS records

Check your [DigitalOcean UI](https://cloud.digitalocean.com/networking/domains) to view the records for your DigitalOcean DNS zone.

Click on the zone for the one created above if a different domain was used.

This should show the external IP address of the service as the A record for your domain.

## Cleanup

Now that we have verified that ExternalDNS will automatically manage DigitalOcean DNS records, we can delete the tutorial's example:

```
$ kubectl delete service -f nginx.yaml
$ kubectl delete service -f externaldns.yaml
```
