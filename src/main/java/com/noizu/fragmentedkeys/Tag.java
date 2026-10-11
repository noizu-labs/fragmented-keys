package com.noizu.fragmentedkeys;


/**
 * This is the Standard Implementation of a KeyGroup.
 * Additional Implementations are:
 * - Database or File Backed KeyGroups - Commonly you will want to persist versioning information on a layer that
 * is less volatile than memcache. Memcache will perform garbage collection and could remove a keygroup that underpins a large set of data.
 * If this keygroup is not persisted to a database or file persistence layer this garbage collection could result in a unnecessarily large swatch of
 * cache in-validations.
 * - The Constant KeyGroup (A keygroup where the version number is static)
 * - The Time Delayed KeyGroup (a keygroup that uses a time delay when incrementing keygroup version. For example you may only want data up-to-date within
 * 10-30 minutes for some uses and within 1-2 minutes for other uses. A time delayed key group tracks multiple versions itself set to expire at different time intervals.).
 * allowing the end user to specify how stale of data they are willing to work with.
 */
public class Tag implements com.noizu.fragmentedkeys.ITag 
{
	public static final String VERSION_SEPERATOR = ":v";
	public static final String INDEX_SEPERATOR = "_";


	/**
	 * tag group
	 * (e.g. user, apple, etc.)
	 */
	protected String tagName;

	/**
	 * instance/index within group
	 * e.g. user:Keith
	 */
	protected String tagInstance;


	/**
	 * tag-instance version.
	 * E.g. user:keith:version5
	 * @var int
	 */
	protected Long version;

	protected String cachePrefix;

	/**
	 * cache handler
	 * @var \NoizuLabs\FragmentedKeys\ICacheHandler
	 */
	com.noizu.fragmentedkeys.ICacheHandler cacheHandler;

	/**
	 * @param string $tag - name of tag/group ("user", "day_of_week", "username", etc)
	 * @param string $instance - Tag instance (unique id of group record in question)
	 * @param int $version - optional group version can be manually set if desired.
	 * @param ICacheHandler $handler - optional cache handler override
	 * @param string $prefix - optional prefix override
	 */		


	public Tag(String tag) {
		this.tagName = tag;
		this.tagInstance = "";
		this.version = null;
		this.cacheHandler = com.noizu.fragmentedkeys.Configuration.getDefaultCacheHandler();
		this.cachePrefix = com.noizu.fragmentedkeys.Configuration.getGlobalPrefix();
	}	

	public Tag(String tag, String instance) {
		this.tagName = tag;
		this.tagInstance = instance;
		this.version = null;
		this.cacheHandler = com.noizu.fragmentedkeys.Configuration.getDefaultCacheHandler();
		this.cachePrefix = com.noizu.fragmentedkeys.Configuration.getGlobalPrefix();
	}

	public Tag(String tag, String instance, Long version) {
		this.tagName = tag;
		this.tagInstance = instance;
		this.version = version;
		this.cacheHandler = com.noizu.fragmentedkeys.Configuration.getDefaultCacheHandler();
		this.cachePrefix = com.noizu.fragmentedkeys.Configuration.getGlobalPrefix();
	}

	public Tag(String tag, String instance, Long version, com.noizu.fragmentedkeys.ICacheHandler handler) {
		this.tagName = tag;
		this.tagInstance = instance;
		this.version = version;
		this.cacheHandler = (handler != null) ? handler : com.noizu.fragmentedkeys.Configuration.getDefaultCacheHandler();
		this.cachePrefix = com.noizu.fragmentedkeys.Configuration.getGlobalPrefix();
	}

	public Tag(String tag, String instance, Long version, ICacheHandler handler, String prefix)
	{		
		this.tagName = tag;
		this.tagInstance = instance;
		this.version = version;		
		this.cacheHandler = (handler != null) ? handler : com.noizu.fragmentedkeys.Configuration.getDefaultCacheHandler();
		this.cachePrefix = (prefix != null) ? prefix : com.noizu.fragmentedkeys.Configuration.getGlobalPrefix();
	}

	/**
	 * This Control function is used to determine if the fragmentedKey that contains this keygroup may multiget memcache for its version value.
	 * For items like Static Keys we set this value to false to reduce strain on the memcache server albeit at increased php processing time to check this value.
	 */
	public Boolean DelegateMemcacheQuery(String group) {
		return (group == cacheHandler.getGroupName()); 
	}

	public ICacheHandler getCacheHandler() {
		return cacheHandler; 
	}

	public void setCacheHandler(ICacheHandler handler) {
		cacheHandler = handler;
	}

	/**
	 * returns the current version value for this group.
	 */
	public Long getTagVersion()
	{
		return (version != null) ? version : _getVersion();
	}

	/**
	 * Modify the version used for this group.
	 * @param int $version
	 * @param bool $update - Update Memcache/Persistant store.
	 */
	public void setTagVersion(Long version, Boolean update)
	{
		this.version = version;
		if(update) {
			_StoreVersion();
		}
	}

	/**
	 * Retrieve this groups tag (Unique GroupName Plus Versioning Info );
	 * @return
	 */
	public String getFullTag() {
		return getTagName() + VERSION_SEPERATOR + getTagVersion();
	}

	/**
	 * Retrieve this groups name. (For example USER_123423);
	 * @return
	 */
	public String getTagName()
	{
		return tagName + INDEX_SEPERATOR + tagInstance + cachePrefix;
	}

	/**
	 * Increment version number.
	 */
	public void Increment()
	{
		if(version == null) {
			_getVersion();
		}
		version += 1;
		_StoreVersion();
	}

	/**
	 * Reset version number. We use a microtime stamp to insure this value is
	 * always unique and will not result in a pull of invalidated data.
	 */
	public void ResetTagVersion()
	{
		version = 0l; //TODO - Time in microseconds 
		_StoreVersion();
	}

	/**
	 * return current version number of tag-instance
	 * @return int
	 */
	protected Long _getVersion()
	{
		if(version == null)
		{
			Long result = (Long) cacheHandler.get(getTagName());
			if(result == null) {
				ResetTagVersion();
			} else {
				version = result;
			}
		}
		return version;
	}

	/**
	 * update version in cache.
	 */
	protected void _StoreVersion()
	{
		cacheHandler.set(getTagName(), version);
	}
}