package com.noizu.fragmentedkeys.tag;

import com.noizu.fragmentedkeys.ICacheHandler;

/**
 * This is the Standard Implementation of a KeyGroup. Versions are updated as soon as they are invalidated.
 */
public class Standard extends com.noizu.fragmentedkeys.Tag
{
	public Standard(String tag) {
		super(tag);
	}
	public Standard(String tag, String instance) {
		super(tag, instance);
	}
	public Standard(String tag, String instance, Long version) {
		super(tag, instance, version);
	}
	public Standard(String tag, String instance, Long version, ICacheHandler handler) {
		super(tag, instance, version, handler);
	}
	public Standard(String tag, String instance, Long version, ICacheHandler handler, String prefix) {
		super(tag, instance, version, handler, prefix);
	}
	
	public Standard(String tag, Integer instance) {
		super(tag, instance.toString());
	}
	public Standard(String tag, Integer instance, Long version) {
		super(tag, instance.toString(), version);
	}
	public Standard(String tag, Integer instance, Long version, ICacheHandler handler) {
		super(tag, instance.toString(), version, handler);
	}
	public Standard(String tag, Integer instance, Long version, ICacheHandler handler, String prefix) {
		super(tag, instance.toString(), version, handler, prefix);
	}
	
	
	public Standard(String tag, Long instance) {
		super(tag, instance.toString());
	}
	public Standard(String tag, Long instance, Long version) {
		super(tag, instance.toString(), version);
	}
	public Standard(String tag, Long instance, Long version, ICacheHandler handler) {
		super(tag, instance.toString(), version, handler);
	}
	public Standard(String tag, Long instance, Long version, ICacheHandler handler, String prefix) {
		super(tag, instance.toString(), version, handler, prefix);
	}	
}