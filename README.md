# podmove
move pods by using the RC/RS adoption 

# move a pod which is controlled by a ReplicaSet or ReplicationController
There are three steps:
 1. Create a clone Pod of the original Pod
 
   The cloned pod has everything except the labels and podName of the original Pod;
   
 2. Delete the original Pod
   
   
 3. Update the new Pod by adding the labels
 
 After a while, the new created Pod will be adopted by the ReplicaSet/ReplicaionController.
 
 
