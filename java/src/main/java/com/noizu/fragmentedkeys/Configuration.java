package com.noizu.fragmentedkeys;


/**
 * Configuration Settings (Dependency Injection)
 */
public class Configuration
{
	public static final Boolean BACKWORDS_COMPATIBLE = true;
	static protected ICacheHandler defaultCacheHandler;
	static protected String globalPrefix;
	
	/**
	 * Specify the default mechanism to use for fetching and retrieving tag values.
	 * @return ICacheHandler
	 */
	public static void setDefaultCacheHandler(ICacheHandler handler){
		defaultCacheHandler = handler;
	}
	
	/**
	 * Specify the default mechanism to use for fetching and retrieving tag values.
	 * @return type
	 */
	 public static ICacheHandler getDefaultCacheHandler() {
		/* May Delete after May 1st 2015 - noizu
		if(!isset(self::$defaultCacheHandler)) {
			global $container;
			if(self::BACKWORDS_COMPATIBLE && isset($container) && isset($container['memcache'])) {
				// Backwards Compatibility
				if(class_exists("Memcache") && is_a($container['memcache'], "Memcache")) {
					self::$defaultCacheHandler = new CacheHandler\Memcache($container['memcache']);
				} else if (class_exists("Memcached") && is_a($container['memcache'], "Memcached")) {
					self::$defaultCacheHandler = new CacheHandler\Memcached($container['memcache']);
				} else {
					trigger_error("Before using NoizuLabs/Fragmented you must set a cache handler using NoizuLabs\FragmentedKeys\Configuration::setDefaultCacheHandler(ICacheHandle \$handler) or use only KeyRing entries with a specified storage mechanism.", E_USER_ERROR);
				}
			} else {
				trigger_error("Before using NoizuLabs/Fragmented you must set a cache handler using NoizuLabs\FragmentedKeys\Configuration::setDefaultCacheHandler(ICacheHandle \$handler) or use only KeyRing entries with a specified storage mechanism.", E_USER_ERROR);
			}
		}
		return self::$defaultCacheHandler;
		*/
		 return defaultCacheHandler;
	}

	/**
	 * Set the global prefix added to all keys by FragementedKeySystem.
	 * @param String $prefix
	 */
	public static void setGlobalPrefix(String prefix)
	{
		globalPrefix = prefix;
	}

	/**
	 * Get the global prefix added to all keys by FragementedKeySystem.
	 * @return string
	 */
	public static String getGlobalPrefix()
	{
		/* May Delete after May 1st 2015 - noizu
		if(!isset(self::$globalPrefix)) {
			// backwards compatibility
			global $container;
			if (self::BACKWORDS_COMPATIBLE) {
				if (isset($container) && isset($container['memcachePrefix'])) {
					self::$globalPrefix = $container['memcachePrefix'];
				} else {
					self::$globalPrefix = "DefaultPrefix";
				}
			} else {
				trigger_error("Before using NoizuLabs/Fragmented you must set a global prefix using NoizuLabs\FragmentedKeys\Configuration::setGlobalPrefix(String \$prefix)", E_USER_WARNING);
			}
		}
		return self::$globalPrefix;
		*/
		return globalPrefix;
	}
}