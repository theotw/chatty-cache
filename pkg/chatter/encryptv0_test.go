/*
 * Copyright (c) The One True Way 2023. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package chatter

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAesCBC(t *testing.T) {
	key0 := makeAesKey("testphrase")
	key1 := makeAesKey("testphrase that is really long")
	key2 := makeAesKey("a")
	//sanity test
	assert.Equal(t, len(key0), len(key1))
	assert.Equal(t, len(key0), len(key2))

	plainText := "Space is big, really big.  If you thought the walk to the corner store was big, its peanuts compared to space"
	cipherText, err := DoAesCBCEncrypt([]byte(plainText), key0)
	assert.Nil(t, err, "should be no errors")
	assert.NotNil(t, cipherText, "should be some data encrypted")
	plainBits2, err2 := DoAesCBCDecrypt(cipherText, key0)
	assert.Nil(t, err2, "should be no decypt errors")
	assert.NotNil(t, plainBits2, "should be some data dectpted")
	plainText2 := string(plainBits2)
	assert.Equal(t, plainText, plainText2, "we should get the same things back")

}
