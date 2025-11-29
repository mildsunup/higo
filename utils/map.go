package utils

// --- Map 操作 ---

// Keys 返回 map 的所有键
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Values 返回 map 的所有值
func Values[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// GetOrDefault 获取值，不存在返回默认值
func GetOrDefault[K comparable, V any](m map[K]V, key K, defaultVal V) V {
	if v, ok := m[key]; ok {
		return v
	}
	return defaultVal
}

// Merge 合并多个 map（后者覆盖前者）
func Merge[K comparable, V any](maps ...map[K]V) map[K]V {
	result := make(map[K]V)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// FilterMap 过滤 map
func FilterMap[K comparable, V any](m map[K]V, predicate func(K, V) bool) map[K]V {
	result := make(map[K]V)
	for k, v := range m {
		if predicate(k, v) {
			result[k] = v
		}
	}
	return result
}

// MapKeys 转换 map 的键
func MapKeys[K1, K2 comparable, V any](m map[K1]V, mapper func(K1) K2) map[K2]V {
	result := make(map[K2]V, len(m))
	for k, v := range m {
		result[mapper(k)] = v
	}
	return result
}

// MapValues 转换 map 的值
func MapValues[K comparable, V1, V2 any](m map[K]V1, mapper func(V1) V2) map[K]V2 {
	result := make(map[K]V2, len(m))
	for k, v := range m {
		result[k] = mapper(v)
	}
	return result
}

// Invert 反转 map（键值互换）
func Invert[K, V comparable](m map[K]V) map[V]K {
	result := make(map[V]K, len(m))
	for k, v := range m {
		result[v] = k
	}
	return result
}

// Pick 选取指定键
func Pick[K comparable, V any](m map[K]V, keys ...K) map[K]V {
	result := make(map[K]V)
	for _, k := range keys {
		if v, ok := m[k]; ok {
			result[k] = v
		}
	}
	return result
}

// Omit 排除指定键
func Omit[K comparable, V any](m map[K]V, keys ...K) map[K]V {
	exclude := make(map[K]struct{}, len(keys))
	for _, k := range keys {
		exclude[k] = struct{}{}
	}

	result := make(map[K]V)
	for k, v := range m {
		if _, ok := exclude[k]; !ok {
			result[k] = v
		}
	}
	return result
}
