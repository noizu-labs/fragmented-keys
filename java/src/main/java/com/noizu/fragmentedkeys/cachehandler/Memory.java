package com.noizu.fragmentedkeys.cachehandler;

import java.util.HashMap;
import java.util.List;
import java.util.Map;


public class Memory implements com.noizu.fragmentedkeys.ICacheHandler  {
	static protected HashMap<String, Object> cache = new HashMap<String, Object>();
	
	public Memory() {
	}
	
	public java.lang.Object get(String key) {
		return cache.containsKey(key) ? cache.get(key) : null;
	}
	
	public String getGroupName() {
		return "Memory";
	}
	
	public void set(String key, java.lang.Object value) {
		set(key, value, 0);		
	}
	
	public void set(String key, java.lang.Object value, Integer period) {
		// TODO Emulate Cache Invalidation. 
		cache.put(key, value);
	}
	
	public Map<String, java.lang.Object> getMulti(List<String> keys) {
		HashMap response = new HashMap<String, java.lang.Object>();
		for(String key: keys) {
			if (cache.containsKey(key)) {
				response.put(key, cache.get(key));
			}
		}
		return response;
	}
}