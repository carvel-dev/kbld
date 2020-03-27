package search

type ObjVisitor struct {
	obj     interface{}
	matcher Matcher
}

func NewObjVisitor(res interface{}, matcher Matcher) ObjVisitor {
	return ObjVisitor{res, matcher}
}

func (kvs ObjVisitor) Visit(visitorFunc func(interface{}) (interface{}, bool)) {
	kvs.visitValues(kvs.obj, visitorFunc)
}

func (kvs ObjVisitor) visitValues(obj interface{}, visitorFunc func(interface{}) (interface{}, bool)) {
	switch typedObj := obj.(type) {
	case map[string]interface{}:
		for k, v := range typedObj {
			if kvs.matcher.Matches(k, v) {
				if newVal, update := visitorFunc(v); update {
					typedObj[k] = newVal
				}
			} else {
				kvs.visitValues(typedObj[k], visitorFunc)
			}
		}

	case map[string]string:
		for k, v := range typedObj {
			if kvs.matcher.Matches(k, v) {
				if newVal, update := visitorFunc(v); update {
					typedObj[k] = newVal.(string)
				}
			} else {
				kvs.visitValues(typedObj[k], visitorFunc)
			}
		}

	case []interface{}:
		for _, o := range typedObj {
			kvs.visitValues(o, visitorFunc)
		}
	}
}
