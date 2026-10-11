package com.noizu.fragmentedkeys.tag;

import com.noizu.fragmentedkeys.ICacheHandler;

/**
 * This is the Constant Implementation of a KeyGroup. Its version never changes.
 */
public class Constant extends com.noizu.fragmentedkeys.Tag
{

	public Constant(String tag) {
		super(tag);
		if (version == null) {
			version = 1l;
		}		
	}
	public Constant(String tag, String instance) {
		super(tag, instance);
		if (version == null) {
			version = 1l;
		}		
	}
	public Constant(String tag, String instance, Long version) {
		super(tag, instance, version);
		if (version == null) {
			version = 1l;
		}		
	}
	public Constant(String tag, String instance, Long version, ICacheHandler handler) {
		super(tag, instance, version, handler);
		if (version == null) {
			version = 1l;
		}		
	}
	public Constant(String tag, String instance, Long version, ICacheHandler handler, String prefix) {
		super(tag, instance, version, handler, prefix);
		if (version == null) {
			version = 1l;
		}		
	}
	
	
	public Constant(String tag, Integer instance) {
		super(tag, instance.toString());
		if (version == null) {
			version = 1l;
		}		
	}
	public Constant(String tag, Integer instance, Long version) {
		super(tag, instance.toString(), version);
		if (version == null) {
			version = 1l;
		}		
	}
	public Constant(String tag, Integer instance, Long version, ICacheHandler handler) {
		super(tag, instance.toString(), version, handler);
		if (version == null) {
			version = 1l;
		}		
	}
	public Constant(String tag, Integer instance, Long version, ICacheHandler handler, String prefix) {
		super(tag, instance.toString(), version, handler, prefix);
		if (version == null) {
			version = 1l;
		}		
	}	

	public Constant(String tag, Long instance) {
		super(tag, instance.toString());
		if (version == null) {
			version = 1l;
		}		
	}
	public Constant(String tag, Long instance, Long version) {
		super(tag, instance.toString(), version);
		if (version == null) {
			version = 1l;
		}		
	}
	public Constant(String tag, Long instance, Long version, ICacheHandler handler) {
		super(tag, instance.toString(), version, handler);
		if (version == null) {
			version = 1l;
		}		
	}
	public Constant(String tag, Long instance, Long version, ICacheHandler handler, String prefix) {
		super(tag, instance.toString(), version, handler, prefix);
		if (version == null) {
			version = 1l;
		}		
	}		
	
	/**
	 * Because we are a constant Tag we never need to have our version updated..
	 * @param string $group
	 */
	public Boolean DelegateMemcacheQuery(String group) {
		return false;
	}
	
	/**
	 * returns the current version value for this group.
	 */
	public Long getTagVersion()
	{
		return version;
	}

	/**
	 * Modify the version used for this group (Does nothing against constant tags)
	 * @param int $version
	 * @param bool $update - Update Memcache/Persistant store.
	 */
	public void setTagVersion(Long version, Boolean update)
	{
		/* Do Nothing, Constant Tag*/
	}

	/**
	 * Retrieve this groups tag (Unique GroupName Plus Versioning Info );
	 */
	public String getFullTag() {
		return getTagName() + VERSION_SEPERATOR + getTagVersion();
	}

	/**
	 * Retrieve this groups name. (For example USER_123423);
	 */
	public String getTagName()
	{
		return tagName + INDEX_SEPERATOR  + tagInstance + cachePrefix;
	}

	/**
	 * Increment version number.
	 */
	public void Increment()
	{
		/*Do Nothing, Constant Tag*/
	}
	/**
	 * Reset version number. We use a microtime stamp to insure this value is
	 * always unique and will not result in a pull of invalidated data.
	 */
	public void ResetTagVersion()
	{
		/*Do Nothing, Constant Tag*/
	}

	/**
	 * return current version number of tag-instance
	 */
	protected Long _getVersion()
	{
		return version;
	}

	/**
	 * update version in cache.
	 */
	protected void _StoreVersion()
	{
		/*Do Nothing, Constant Tag*/
	}
}