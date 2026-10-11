package com.noizu.fragmentedkeys;

import java.util.List;
import java.util.Map;

/**
 * The Key Group is the basic tag + version_info node which our fragmented
 * key system relies on.
 */
public interface ICacheHandler
{
	/**
	 * Constant value used to differentiate between handles when performing multi-gets inside of Key class.
	 */
	public String getGroupName();

	/**
	 * fetch value from cache
	 */
	public java.lang.Object get(String key);

	/**
	 * set value in cache.
	 */
	public void set(String key, java.lang.Object value);
	public void set(String key, java.lang.Object value, Integer period);

	/**
	 * fetch multiple values from cache.
	 */	
	public Map<String, java.lang.Object> getMulti(List<String> keys);	
}