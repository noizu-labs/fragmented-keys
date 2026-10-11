package com.noizu.fragmentedkeys.key;

import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import org.apache.commons.lang3.StringUtils;

import com.noizu.fragmentedkeys.Helper;

/**
 * This is a standard implementation of a fragmented key.
 */
public class Standard implements com.noizu.fragmentedkeys.IKey  {
	public static final String TAG_SEPERATOR = ":t";
	public static final String INDEX_SEPERATOR = "_";

	protected String key;
	protected String groupId;
	protected Map<String, com.noizu.fragmentedkeys.ITag>tags;


	public Standard(String key, List<com.noizu.fragmentedkeys.ITag> tags)
	{
		this.key = key;		
		this.groupId = "";

		this.tags = new HashMap<String, com.noizu.fragmentedkeys.ITag>();
		for(com.noizu.fragmentedkeys.ITag tag : tags) {
			this.tags.put(tag.getTagName(), tag);
		}
	}
	
	
	public Standard(String key, List<com.noizu.fragmentedkeys.ITag> tags, String groupId)
	{
		this.key = key;		
		this.groupId = groupId;

		this.tags = new HashMap<String, com.noizu.fragmentedkeys.ITag>();
		for(com.noizu.fragmentedkeys.ITag tag : tags) {
			this.tags.put(tag.getTagName(), tag);
		}
	}


	public void AddKeyGroup(com.noizu.fragmentedkeys.ITag tag) {
		this.tags.put(tag.getTagName(), tag);
	}


	public String getKey() {
		return getKey(true);
	}

	public String getKey(Boolean hash)
	{
		return getKeyStr(hash);
	}

	/**
	 * calculate composite key
	 * @param bool $hash use true to return md5 (memcache friendly) key or use false to return raw key for visual inspection.
	 * @return string
	 */
	public String getKeyStr() {
		return getKeyStr(true);
	}

	public String getKeyStr(Boolean hash)
	{		
		String compositeKey = key +  INDEX_SEPERATOR + groupId  + TAG_SEPERATOR  + StringUtils.join(gatherTags(), TAG_SEPERATOR);
		if(hash) {
			// TODO MD5 STRING
			return Helper.MD5(compositeKey);
		} else {
			return compositeKey;
		}
	}

	/**
	 * Bulk Fetch tag-instance versions.
	 * While it would be architecturally cleaner to gather group versions from a KeyGroup function call
	 * the use of memcache/apc multiget helps us avoid some bottle-necking produced by the increased number of key-versions we
	 * need to look-up when using this fragmented key system.
	 */
	protected void GatherGroupVersions() {

		HashMap<String, ArrayList<String>> tagHandlers = new HashMap<String, ArrayList<String>>();
		for (Map.Entry<String, com.noizu.fragmentedkeys.ITag> tagEntry : tags.entrySet())
		{
			String group = tagEntry.getValue().getCacheHandler().getGroupName();
			if (!tagHandlers.containsKey(group)) {
				tagHandlers.put(group, new ArrayList<String>());
			}
			tagHandlers.get(group).add(tagEntry.getKey());			
		}

		
		//HashMap<String, Long> tagVersions = new HashMap<String, Long>();
		HashMap<String, Object> tagMultiGetVersions = new HashMap<String, Object>();
		
		for (Map.Entry<String,ArrayList<String>> tagHandlerEntry : tagHandlers.entrySet())
		{
			ArrayList<String> multiFetchTags = new ArrayList<String>(); 
			for(String delegateTag: tagHandlerEntry.getValue()) {
				if(this.tags.get(delegateTag).DelegateMemcacheQuery(tagHandlerEntry.getKey()) == false) {
					// TODO - DO NOT MULTI_FETCH THIS TAG					
				} else {
					multiFetchTags.add(delegateTag);
				}
			}
			if(multiFetchTags.size() > 0) {
				com.noizu.fragmentedkeys.ICacheHandler handler = this.tags.get(multiFetchTags.get(0)).getCacheHandler();
				tagMultiGetVersions.putAll(handler.getMulti(multiFetchTags));				
			}
		}

		for (Map.Entry<String, com.noizu.fragmentedkeys.ITag> tagEntry:  this.tags.entrySet()) {
			if (tagMultiGetVersions.containsKey(tagEntry.getKey())) {
				tagEntry.getValue().setTagVersion((Long)tagMultiGetVersions.get(tagEntry.getKey()), false);
			} else {
				tagEntry.getValue().ResetTagVersion();
			}
		}
	}

	/**
	 *
	 * @return array tag strings of all tags included within this key.
	 */
	protected List<String> gatherTags()
	{
		GatherGroupVersions();
		ArrayList<String> tagList = new ArrayList<String>();		
		for (Map.Entry<String, com.noizu.fragmentedkeys.ITag> entry : tags.entrySet())
		{
			tagList.add(entry.getValue().getFullTag());
		}
		
		// Sort first for consistent cache keys.
		Collections.sort(tagList);
		return tagList;
	}

	public String toString() {
		return getKey();
	}
}
