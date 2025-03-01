# **IIoT Data Processing with GKE and Pub/Sub**

This project sets up an **Industrial IoT (IIoT) data processing pipeline** using **Google Cloud Platform (GCP)**, **Kubernetes (GKE)**, **Pub/Sub**, and **Golang**. 

## **📌 Overview**
- **Simulated Sensors (VMs)** publish data to **Pub/Sub**.
- **GKE Worker Pods** consume the events and process the data.

---

## **⚡ 1. Prerequisites**
### ✅ Install Required Tools
Make sure you have the following installed:

1. **Google Cloud CLI** → [Install](https://cloud.google.com/sdk/docs/install) (To interact with GCP services from the terminal)
2. **Kubectl** → [Install](https://kubernetes.io/docs/tasks/tools/install-kubectl/) (To manage Kubernetes clusters)
3. **Golang** → [Install](https://go.dev/doc/install) (For worker application development)

### ✅ Set Up Environment Variables in Windows CMD
Instead of hardcoding project details, use environment variables:
```cmd
set PROJECT_ID=your-gcp-project-id
set ZONE=you-zone
set SUBSCRIPTION_ID=your-subscription-id
```

This makes commands reusable and reduces manual errors.

### ✅ Set Up GCP Project
```cmd
gcloud auth login
gcloud config set project %PROJECT_ID%
```

This authenticates you and ensures commands are executed in the correct project.

---

## **🏗 2. Creating Virtual Machines (Sensors)**
### **2.1 Create a VPC for the Project**
```cmd
gcloud compute networks create iiot-vpc --subnet-mode=auto
```

We will use a custom one to isolate network resources for security.

### **2.2 Create Sensor VMs**
```cmd
gcloud compute instances create sensor1 sensor2 ^
  --zone=%ZONE% ^
  --machine-type=e2-micro ^
  --image-family=debian-11 ^
  --image-project=debian-cloud ^
  --network=iiot-vpc ^
  --tags=sensor
```

These VMs simulate IoT sensors sending data.

### **2.3 Create and Assign Custom Service Accounts for Sensors**

#### **2.3.1 Create Service Accounts for Sensors**
```cmd
gcloud iam service-accounts create sensor-publisher ^
  --display-name "Sensor Publisher Service Account"
```

This allows sensors to authenticate and publish messages to Pub/Sub.

#### **2.3.2 Grant Pub/Sub Publisher Role**
```cmd
gcloud projects add-iam-policy-binding %PROJECT_ID% ^
  --member "serviceAccount:sensor-publisher@%PROJECT_ID%.iam.gserviceaccount.com" ^
  --role roles/pubsub.publisher
```

This gives sensors the necessary permissions to publish data.

#### **2.3.3 Attach the Service Account to Sensor VMs**
```cmd
gcloud compute instances set-service-account sensor1 ^
  --zone=%ZONE% ^
  --service-account=sensor-publisher@%PROJECT_ID%.iam.gserviceaccount.com ^
  --scopes=https://www.googleapis.com/auth/cloud-platform
```

This binds the service account to the VMs, so they can authenticate automatically. Please, make sure your VM is stopped before changing scopes otherwise you will not able to set the correct scopes. An alternative to avoid doing this is to set up the correct scopes from the beginning.

---

## **📢 3. Setting Up Pub/Sub**
### **3.1 Create a Pub/Sub Topic**
```cmd
gcloud pubsub topics create sensor-data
```

A topic acts as a communication channel for event messages.

### **3.2 Create a Subscription**
```cmd
gcloud pubsub subscriptions create %SUBSCRIPTION_ID% ^
  --topic=sensor-data
```

Subscriptions allow the consumer (GKE worker) to receive messages.

---

## **🌍 4. Setting Up GKE Cluster**
### **4.1 Enable Kubernetes Engine API**
```cmd
gcloud services enable container.googleapis.com
```

This enables the Kubernetes Engine API, which is required to create clusters (if not already done).

### **4.2 Create the GKE Cluster**
```cmd
gcloud container clusters create iiot-cluster ^
  --zone=%ZONE% ^
  --num-nodes=2 ^
  --enable-ip-alias ^
  --workload-pool=%PROJECT_ID%.svc.id.goog
```

Creates a Kubernetes cluster to run our worker service.

### **4.3 Connect to the Cluster**
```cmd
gcloud container clusters get-credentials iiot-cluster --zone=%ZONE%
```

This allows `kubectl` to manage the cluster.

---

## **🔐 5. Setting Up IAM & Workload Identity**
### **5.1 Create a GCP Service Account for GKE**
```cmd
gcloud iam service-accounts create gke-processor ^
  --display-name "GKE Processor Service Account"
```

This account allows Kubernetes workloads to authenticate with GCP services.

### **5.2 Grant Permissions**
```cmd
gcloud projects add-iam-policy-binding %PROJECT_ID% ^
  --member "serviceAccount:gke-processor@%PROJECT_ID%.iam.gserviceaccount.com" ^
  --role roles/pubsub.subscriber
```

Allows the worker pods to read messages from Pub/Sub.

### **5.3 Create a Kubernetes ServiceAccount (KSA)**
```cmd
kubectl create serviceaccount gke-processor
```

Kubernetes workloads need a service account to interact with cloud services.

### **5.4 Link KSA to GCP IAM Service Account**
```cmd
gcloud iam service-accounts add-iam-policy-binding gke-processor@%PROJECT_ID%.iam.gserviceaccount.com ^
  --role roles/iam.workloadIdentityUser ^
  --member "serviceAccount:%PROJECT_ID%.svc.id.goog[default/gke-processor]"
```

This links the Kubernetes Service Account (KSA) with the GCP IAM service account, allowing authentication without storing credentials in pods.

### **5.5 Annotate Kubernetes ServiceAccount**
```cmd
kubectl annotate serviceaccount gke-processor ^
    iam.gke.io/gcp-service-account=gke-processor@%PROJECT_ID%.iam.gserviceaccount.com
```

This tells Kubernetes which GCP service account the worker pods should use.

---

## **👷 6. Deploying the Worker to GKE**
### **6.1 Create a ConfigMap for Environment Variables**
```cmd
kubectl create configmap worker-config ^
  --from-literal=PROJECT_ID=%PROJECT_ID% ^
  --from-literal=SUBSCRIPTION_ID=%SUBSCRIPTION_ID%
```

This stores configuration values in Kubernetes so they are not hardcoded.

### **6.2 Deploy the Worker Pod**
```cmd
kubectl apply -f worker-deployment.yaml
```

This starts the worker process that listens for Pub/Sub messages.

---

## **📡 7. Testing the System**
### **7.1 Publish a Test Event**
```cmd
gcloud pubsub topics publish sensor-data --message='{"sensor_id": "sensor1", "temperature": 22.5}'
```

Simulates a sensor sending data.

### **7.2 Check Worker Logs**
```cmd
kubectl logs -l app=worker --tail=50 --follow
```

Verifies that the worker is receiving and processing events.

---

