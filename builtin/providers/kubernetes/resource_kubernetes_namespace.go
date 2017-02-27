package kubernetes

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"k8s.io/kubernetes/pkg/api/errors"
	api "k8s.io/kubernetes/pkg/api/v1"
	kubernetes "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_5"
)

func resourceKubernetesNamespace() *schema.Resource {
	return &schema.Resource{
		Create: resourceKubernetesNamespaceCreate,
		Read:   resourceKubernetesNamespaceRead,
		Update: resourceKubernetesNamespaceUpdate,
		Delete: resourceKubernetesNamespaceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"metadata": &schema.Schema{
				Type:        schema.TypeList,
				Description: "Standard object's metadata. More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata",
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"annotations": &schema.Schema{
							Type:         schema.TypeMap,
							Description:  "An unstructured key value map stored with a resource that may be set by external tools to store and retrieve arbitrary metadata. They are not queryable and should be preserved when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations",
							Optional:     true,
							ValidateFunc: validateAnnotations,
						},
						"generate_name": &schema.Schema{
							Type:          schema.TypeString,
							Description:   "Prefix, used by the server, to generate a unique name ONLY IF the `name` field has not been provided. This value will also be combined with a unique suffix. The provided value has the same validation rules as the `name` field, and may be truncated by the length of the suffix required to make the value unique on the server. More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#idempotency",
							Optional:      true,
							ForceNew:      true,
							ValidateFunc:  validateGenerateName,
							ConflictsWith: []string{"metadata.name"},
						},
						"generation": &schema.Schema{
							Type:        schema.TypeInt,
							Description: "A sequence number representing a specific generation of the desired state.",
							Computed:    true,
						},
						"labels": &schema.Schema{
							Type:         schema.TypeMap,
							Description:  "Map of string keys and values that can be used to organize and categorize (scope and select) objects. May match selectors of replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels",
							Optional:     true,
							ValidateFunc: validateLabels,
						},
						"name": &schema.Schema{
							Type:          schema.TypeString,
							Description:   "Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names",
							Optional:      true,
							ForceNew:      true,
							Computed:      true,
							ValidateFunc:  validateName,
							ConflictsWith: []string{"metadata.generate_name"},
						},
						"resource_version": &schema.Schema{
							Type:        schema.TypeString,
							Description: "An opaque value that represents the internal version of this object that can be used by clients to determine when objects have changed. May be used for optimistic concurrency, change detection, and the watch operation on a resource or set of resources. Clients must treat these values as opaque and passed unmodified back to the server. They may only be valid for a particular resource or set of resources. More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#concurrency-control-and-consistency",
							Computed:    true,
						},
						"self_link": &schema.Schema{
							Type:        schema.TypeString,
							Description: "A URL representing this object.",
							Computed:    true,
						},
						"uid": &schema.Schema{
							Type:        schema.TypeString,
							Description: "The unique in time and space value for this object. More info: http://kubernetes.io/docs/user-guide/identifiers#uids",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func resourceKubernetesNamespaceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetes.Clientset)

	metadata := expandMetadata(d.Get("metadata").([]interface{}))
	namespace := api.Namespace{
		ObjectMeta: metadata,
	}
	log.Printf("[INFO] Creating new namespace: %#v", namespace)
	out, err := conn.CoreV1().Namespaces().Create(&namespace)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Submitted new namespace: %#v", out)
	d.SetId(out.Name)

	return resourceKubernetesNamespaceRead(d, meta)
}

func resourceKubernetesNamespaceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetes.Clientset)

	name := d.Id()
	log.Printf("[INFO] Reading namespace %s", name)
	namespace, err := conn.CoreV1().Namespaces().Get(name)
	if err != nil {
		log.Printf("Received error: %#v", err)
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			log.Printf("[WARN] Removing namespace %s (it is gone)", name)
			d.SetId("")
			return nil
		}
		return err
	}
	log.Printf("[INFO] Received namespace: %#v", namespace)
	err = d.Set("metadata", flattenMetadata(namespace.ObjectMeta))
	if err != nil {
		return err
	}

	return nil
}

func resourceKubernetesNamespaceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetes.Clientset)

	metadata := expandMetadata(d.Get("metadata").([]interface{}))
	// This is necessary in case the name is generated
	metadata.Name = d.Id()

	namespace := api.Namespace{
		ObjectMeta: metadata,
	}
	log.Printf("[INFO] Updating namespace: %#v", namespace)
	out, err := conn.CoreV1().Namespaces().Update(&namespace)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Submitted updated namespace: %#v", out)
	d.SetId(out.Name)

	return resourceKubernetesNamespaceRead(d, meta)
}

func resourceKubernetesNamespaceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetes.Clientset)

	name := d.Id()
	log.Printf("[INFO] Deleting namespace: %#v", name)
	err := conn.CoreV1().Namespaces().Delete(name, &api.DeleteOptions{})
	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Target:  []string{},
		Pending: []string{"Terminating"},
		Timeout: 5 * time.Minute,
		Refresh: func() (interface{}, string, error) {
			out, err := conn.CoreV1().Namespaces().Get(name)
			if err != nil {
				if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
					return nil, "", nil
				}
				log.Printf("[ERROR] Received error: %#v", err)
				return out, "Error", err
			}

			statusPhase := fmt.Sprintf("%v", out.Status.Phase)
			log.Printf("[DEBUG] Namespace %s status received: %#v", out.Name, statusPhase)
			return out, statusPhase, nil
		},
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return err
	}
	log.Printf("[INFO] Namespace %s deleted", name)

	d.SetId("")
	return nil
}