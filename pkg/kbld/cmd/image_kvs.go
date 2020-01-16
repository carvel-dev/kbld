package cmd

type ImageKVs struct {
	Resource interface{}
	Keys     []string
}

func (kvs ImageKVs) Visit(visitorFunc func(interface{}) (interface{}, bool)) {
	for _, key := range kvs.Keys {
		kvs.visitValues(kvs.Resource, key, visitorFunc)
	}
}

func (kvs ImageKVs) visitValues(obj interface{}, key string, visitorFunc func(interface{}) (interface{}, bool)) {
	switch typedObj := obj.(type) {
	case map[string]interface{}:
		for k, v := range typedObj {
			if k == key {
				if newVal, update := visitorFunc(v); update {
					typedObj[k] = newVal
				}
			} else {
				kvs.visitValues(typedObj[k], key, visitorFunc)
			}
		}

	case map[string]string:
		for k, v := range typedObj {
			if k == key {
				if newVal, update := visitorFunc(v); update {
					typedObj[k] = newVal.(string)
				}
			} else {
				kvs.visitValues(typedObj[k], key, visitorFunc)
			}
		}

	case []interface{}:
		for _, o := range typedObj {
			kvs.visitValues(o, key, visitorFunc)
		}
	}
}
