package com.noizu.fragmentedkeys;

/**
 * The Key Group is the basic tag + version_info node which our fragmented
 * key system relies on.
 */
public interface ITag
{	
	/**
	 * get current Tag-Instance Version.
	 */
	public Long getTagVersion();
	
	/**
	 * get cache handler;
	 */
	public com.noizu.fragmentedkeys.ICacheHandler getCacheHandler();
	
	/**
	 * Set Tag-Instance Version to specific value.
	 */
	public void setTagVersion(Long version, Boolean update);
	
	/**
	 * Reset version associated with Tag-Instance
	 */
	public void ResetTagVersion();
	
	/**
	 * Tag Name, Instance & Version
	 */
	public String getFullTag();
	
	/*
	 * get Tag-Instance
	 */
	public String getTagName();
	
	/**
	 * Increment Tag-Instance Version.
	 */
	public void Increment();
	
	/**
	 * Allow Upstream Multi-Get for Performance Reasons
	 */
	public Boolean DelegateMemcacheQuery(String group);
}