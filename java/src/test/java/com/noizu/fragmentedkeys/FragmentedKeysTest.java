package com.noizu.fragmentedkeys;

import static org.junit.Assert.*; 

import java.util.ArrayList;

import org.junit.*;

import com.noizu.fragmentedkeys.*;

/**
 * Unit test for simple App.
 */
public class FragmentedKeysTest 
{
	public static final Integer ONE_TICK = 1;
	private String tagNameA;
	private String tagNameB;
	private Integer tagNameAEntityOne;
	private Integer tagNameAEntityTwo;
	private com.noizu.fragmentedkeys.ICacheHandler memoryHandler;

	//==================================================================================================================
	// SetUp/Teardown
	//==================================================================================================================
	@Before
	public void setUp()
	{		
		Long time = System.currentTimeMillis();
		tagNameA = "TagA_" + time;
		tagNameB = "TagB_" + time;
		tagNameAEntityOne = 1;
		tagNameAEntityTwo = 2;

		com.noizu.fragmentedkeys.Configuration.setDefaultCacheHandler(new com.noizu.fragmentedkeys.cachehandler.Memory());
		com.noizu.fragmentedkeys.Configuration.setGlobalPrefix("Testing");
	}
	//==================================================================================================================
	// Helper Methods
	//==================================================================================================================
	private void WaitForClockTick()
	{
		try {
			Thread.sleep(ONE_TICK * 1000);
		} catch (InterruptedException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();			
		}
	}

	//==================================================================================================================
	// Tests
	//==================================================================================================================
	//-----------------------------
	// Tests - tags
	//-----------------------------
	@Test
	public void StandardTagShouldReturnTheSameValueIfIncrementHasNotBeenCalled()
	{
		//Todo simplify constructor to avoid need to always provide version, handler & prefix. 
		//E.g. leverage global config. 
		ITag tag = new com.noizu.fragmentedkeys.tag.Standard(tagNameA, tagNameAEntityOne);
		String version1 = tag.getFullTag();
		this.WaitForClockTick();
		ITag tag2 = new com.noizu.fragmentedkeys.tag.Standard(tagNameA, tagNameAEntityOne);
		
		String version2 = tag2.getFullTag();
		assertEquals(version1, version2);
	}


	@Test
	public void StandardTagShouldReturnDifferentValuesIfIncrementHasBeenCalled()
	{
		ITag tag = new com.noizu.fragmentedkeys.tag.Standard(tagNameA, tagNameAEntityOne);
		String version1 = tag.getFullTag();
		tag.Increment();
		ITag tag2 = new com.noizu.fragmentedkeys.tag.Standard(tagNameA, tagNameAEntityOne);
	    String version2 = tag2.getFullTag();
		assertFalse(version1 == version2);
	}
	
	@Test
	public void StandardTagsShouldReturnDifferentValuesForDifferentEntities()
	{
		ITag tag = new com.noizu.fragmentedkeys.tag.Standard(tagNameA, tagNameAEntityOne);
		ITag tag2 = new com.noizu.fragmentedkeys.tag.Standard(tagNameA, tagNameAEntityTwo);
		assertFalse(tag.getFullTag() == tag2.getFullTag());
	}

	@Test
	public void StandardTagsShouldReturnDifferentValuesForDifferentTags()
	{
		ITag tag = new com.noizu.fragmentedkeys.tag.Standard(tagNameA, tagNameAEntityOne);
		ITag tag2 = new com.noizu.fragmentedkeys.tag.Standard(tagNameB, tagNameAEntityOne);
		assertFalse(tag.getFullTag() == tag2.getFullTag());
	}
	@Test
	public void CallingIncrementVersionShouldChangeTheVersionOnStandardTagInstances()
	{
		ITag tag = new com.noizu.fragmentedkeys.tag.Standard(tagNameA, tagNameAEntityOne);
		Long version = tag.getTagVersion();
		tag.Increment();
		Long version2 = tag.getTagVersion();
		assertFalse(version ==version2);
	}
	@Test
	public void IncrementingOneTagGroupInstanceShouldNotChangeTheVersionOfOtherTAgInstancePairs()
	{
		ITag tag = new com.noizu.fragmentedkeys.tag.Standard(tagNameA, tagNameAEntityOne);
		ITag tag2 = new com.noizu.fragmentedkeys.tag.Standard(tagNameA, tagNameAEntityTwo);
		String version = tag.getFullTag();
		tag2.Increment();
		assertEquals(version, tag.getFullTag());
	}
	@Test
	public void IncrementingOneTagEntityShouldNotChangeVersionOfDifferentTagForTheSameIdValue()
	{
		ITag tag = new com.noizu.fragmentedkeys.tag.Standard(tagNameA, tagNameAEntityOne);
		ITag tag2 = new com.noizu.fragmentedkeys.tag.Standard(tagNameB, tagNameAEntityOne);
		String version = tag.getFullTag();
		tag2.Increment();
		assertEquals(version, tag.getFullTag());
	}

	

	//-----------------------------
	// Tests - keys
	//-----------------------------
	@Test
	public void GetKeyShouldCorrectlyPullVersionsFromAllTags()
	{
		ITag tag = new com.noizu.fragmentedkeys.tag.Standard(tagNameA, tagNameAEntityOne);
		ITag tag2 = new com.noizu.fragmentedkeys.tag.Standard(tagNameB, tagNameAEntityOne);
		
		ArrayList<ITag> tags = new ArrayList<ITag>();
		tags.add(tag);
		tags.add(tag2);
		IKey theKey = new com.noizu.fragmentedkeys.key.Standard("ThisIsAKey", tags);
		String expected = Helper.MD5("ThisIsAKey_:t" + tag.getFullTag() + ":t" + tag2.getFullTag());
		assertEquals(expected, theKey.getKey());
	}
	
	@Test
	public void GetKeyShouldWorkWithASingleTag()
	{
		ITag tag = new com.noizu.fragmentedkeys.tag.Standard(tagNameA, tagNameAEntityOne);
		
		ArrayList<ITag> tags = new ArrayList<ITag>();
		tags.add(tag);
		IKey theKey = new com.noizu.fragmentedkeys.key.Standard("ThisIsAKey", tags);	
		String expected = Helper.MD5("ThisIsAKey_:t" + tag.getFullTag());
		assertEquals(expected, theKey.getKey());
	}
	@Test
	public void GetKeyShouldReturnSameValueOnSubsequentCallsIfTagsAreNotIncremented()
	{
		ITag tag = new com.noizu.fragmentedkeys.tag.Standard(tagNameA, tagNameAEntityOne);
		ITag tag2 = new com.noizu.fragmentedkeys.tag.Standard(tagNameB, tagNameAEntityOne);
		
		ArrayList<ITag> tags = new ArrayList<ITag>();
		tags.add(tag);
		tags.add(tag2);
		IKey theKey = new com.noizu.fragmentedkeys.key.Standard("ThisIsAKey", tags);		
					
		String firstKey = theKey.getKey(false);
		this.WaitForClockTick();
		
		tags = new ArrayList<ITag>();
		tags.add(tag);
		tags.add(tag2);
		theKey = new com.noizu.fragmentedkeys.key.Standard("ThisIsAKey", tags);
		String secondKey = theKey.getKey(false);
		assertEquals(firstKey, secondKey);
	}

	@Test
	public void GetKeyShouldReturnDifferentValuesIfTagsAreIncremented()
	{
		ITag tag = new com.noizu.fragmentedkeys.tag.Standard(tagNameA, tagNameAEntityOne);
		ITag tag2 = new com.noizu.fragmentedkeys.tag.Standard(tagNameB, tagNameAEntityOne);
		
		ArrayList<ITag> tags = new ArrayList<ITag>();
		tags.add(tag);
		tags.add(tag2);
		IKey theKey = new com.noizu.fragmentedkeys.key.Standard("ThisIsAKey", tags);		
		String firstKey = theKey.getKey(false);
		
		
		
		ITag tag2b = new com.noizu.fragmentedkeys.tag.Standard(tagNameB, tagNameAEntityOne);
		tag2b.Increment();
		

		tags = new ArrayList<ITag>();
		tags.add(tag);
		tags.add(tag2);
		theKey = new com.noizu.fragmentedkeys.key.Standard("ThisIsAKey", tags);
		String secondKey = theKey.getKey(false);
		assertFalse(firstKey == secondKey);
	}
	
	@Test @Ignore
	public void GetKeyShouldBeAbleToProcessGroupsOfTagsWithDifferentCacheHandlersAndReturnTheSameKeyIfNotIncremented()
	{
		/*
		if(is_null(this.apcHandler)) {
			this.markTestSkipped('Command Line APC must be enabled to execute this test');
		}
		tag = new Tag\Standard(this.tagNameA, this.tagNameAEntityOne, null, this.apcHandler);
		ITag tag2 = new com.noizu.fragmentedkeys.tag.Standard(tagNameB, tagNameAEntityOne);
		theKey = new Key\Standard("ThisIsAKey", array(tag,tag2));
		String firstKey = theKey.getKey(false);
		this.WaitForClockTick();
		tag = new Tag\Standard(this.tagNameA, this.tagNameAEntityOne, null, this.apcHandler);
		theKey = new Key\Standard("ThisIsAKey", array(tag,tag2));
		secondKey = theKey.getKey(false);
		assertEquals(firstKey, secondKey);
		*/
	}

	@Test @Ignore
	public void GetKeyShouldBeAbleToProcessGroupsOfTagsWithDifferentCacheHandlersAndReturnADifferentKeyIfIncremented()
	{
		/*
		if(is_null(this.apcHandler)) {
			this.markTestSkipped('Command Line APC must be enabled to execute this test');
		}
		ITag tag = new com.noizu.fragmentedkeys.tag.Standard(tagNameA, tagNameAEntityOne);
		tag2 = new Tag\Standard(this.tagNameB, this.tagNameAEntityOne, null, this.apcHandler);
		theKey = new Key\Standard("ThisIsAKey", array(tag,tag2));
		String firstKey = theKey.getKey(false);
		tag2b = new Tag\Standard(this.tagNameB, this.tagNameAEntityOne, null, this.apcHandler);
		tag2b.Increment();
		tag2b.Increment();
		theKey = new Key\Standard("ThisIsAKey", array(tag,tag2));
		secondKey = theKey.getKey(false);
		assertFlase(firstKey == secondKey);
		*/
	}

	@Test
	public void AKeyShouldRemainConstantWhenUsingOnlyStaticTags()
	{
		ITag tag = new com.noizu.fragmentedkeys.tag.Constant(tagNameA, tagNameAEntityOne, 5L);
		ITag tag2 = new com.noizu.fragmentedkeys.tag.Constant(tagNameB, tagNameAEntityOne, 8L);
		ArrayList<ITag> tags = new ArrayList<ITag>();
		tags.add(tag);
		tags.add(tag2);
		IKey theKey = new com.noizu.fragmentedkeys.key.Standard("ThisIsAKey", tags);		

		
		
		String firstKey = theKey.getKey(false);
		tag.Increment();
		tag2.Increment();
		
		tags = new ArrayList<ITag>();
		tags.add(tag);
		tags.add(tag2);
		theKey = new com.noizu.fragmentedkeys.key.Standard("ThisIsAKey", tags);				
		String secondKey = theKey.getKey(false);
		assertEquals(firstKey, secondKey);
	}
	//-----------------------------
	// Tests - key rings
	//-----------------------------
	@Test @Ignore 
	public void KeyDefinedUsingKeyRingShouldMatchEquivelentManuallyConstructedKey()
	{
		/*
		// Define Key
		cacheHandlers = array(
				'memcache' => new \NoizuLabs\FragmentedKeys\CacheHandler\Memcached(this.container['memcache']),
				'memory' => new \NoizuLabs\FragmentedKeys\CacheHandler\Memory()
				);
		globalOptions = array(
				'type' => 'standard'
				);
		tagOptions = array(
				this.tagNameB => array('type' => 'constant', 'version' => 5)
				);
		ring = new FragmentedKeys\KeyRing(globalOptions, tagOptions, 'memcache', cacheHandlers);
		ring.DefineKey("Users", array(this.tagNameA, array('tag' => this.tagNameB , 'cacheHandler' => 'memory', 'version' => null, 'type'=>'standard'), this.tagNameB));
		ITag tag = new com.noizu.fragmentedkeys.tag.Standard(tagNameA, tagNameAEntityOne);
		tag2 = new Tag\Standard(this.tagNameB, this.tagNameAEntityOne, null, new \NoizuLabs\FragmentedKeys\CacheHandler\Memory());
		tag3 = new Tag\Constant(this.tagNameB, this.tagNameAEntityTwo, 5);
		key1 = new Key\Standard("Users", array(tag,tag2,tag3));
		String firstKey = key1.getKey(false);
		key2 = ring.getUsersKeyObj(this.tagNameAEntityOne, this.tagNameAEntityOne, this.tagNameAEntityTwo);
		secondKey = key2.getKey(false);
		assertEquals(firstKey, secondKey);
		*/
	}	
}