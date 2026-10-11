package com.noizu.fragmentedkeys;

import java.util.List;
import java.util.Map;

/**
 * The various key groups are merged together to form our final cache key.
 */
public interface IKeyRing
{

	public void setTagOptions(String tag, Map<String,String> options) ;

	public void setGlobalOptions(Map<String,String> options);

	public Map<String,String> getTagOptions(String tag, List<String> options);

	public Map<String,String> getGlobalOptions();

	public void DefineKey(String key,  Map<String, Object> params,  Map<String, Object> globals);
	
}