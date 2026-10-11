package com.noizu.fragmentedkeys;

/**
 * The various key groups are merged together to form our final cache key.
 */
public interface IKey  {
	/*
	 * Append Tag to Key.
	 */
	public void AddKeyGroup(com.noizu.fragmentedkeys.ITag tag);
	
	/*
	 * Retrieve Key
	 */
	public String getKey();	
	public String getKey(Boolean hash);
}