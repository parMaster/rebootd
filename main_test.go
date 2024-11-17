package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_testConnection(t *testing.T) {
	// multiple URLs
	err := testConnection("https://google.com/,https://cloudflare.com/,https://amazon.com/")
	assert.Nil(t, err)

	// single URL
	err = testConnection("https://google.com")
	assert.Nil(t, err)

	// one invalid URL
	err = testConnection("https://cloudflare.com/,https://amazon.com/invalid")
	assert.Nil(t, err)

	// comma sepatared URLs, with spaces
	err = testConnection("https://cloudflare.com/,    https://amazon.com/invalid")
	assert.Nil(t, err)

	// all invalid URLs
	err = testConnection("https://aasdasdasdasdasdlkdajflaskdjflksjdlfkjslkdjflakjdlfkjasd.com,https://kjahdkfjhadskjfhaksdjhflkajdshklasjdhfkljvnzcxvkzxcnvmzxcnvlkjdsflkajsndlkcajndczxc.com")
	assert.NotNil(t, err)

	// empty URL
	err = testConnection("")
	assert.NotNil(t, err)

}
