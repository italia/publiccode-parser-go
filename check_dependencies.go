package publiccode

func (p *Parser) checkDependencies(key string, value map[interface{}]interface{}) ([]Dependency, error) {
	var deps []Dependency
	for _, v := range value {
		var dep Dependency
		for k, val := range v.(map[interface{}]interface{}) {
			if k.(string) == "name" {
				dep.Name = val.(string)
			} else if k.(string) == "optional" {
				var ok bool
				if dep.Optional, ok = val.(bool); !ok {
					return nil, newErrorInvalidValue(key, "invalid type for key '%s', boolean expected", k)
				}
			} else if k.(string) == "version" {
				dep.Version = val.(string)
			} else if k.(string) == "versionMin" {
				dep.VersionMin = val.(string)
			} else if k.(string) == "versionMax" {
				dep.VersionMax = val.(string)
			} else {
				return nil, newErrorInvalidValue(key, "unexpected key '%s'", k)
			}
		}
		if dep.Name == "" {
			return nil, newErrorInvalidValue(key, "missing mandatory key 'name'")
		}

		deps = append(deps, dep)
	}
	return deps, nil
}
