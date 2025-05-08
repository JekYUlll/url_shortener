package shortcode

import "math/rand"

type ShortCodeGeneratorImpl struct {
	length int
}

func NewShortCodeGeneratorImpl(length int) *ShortCodeGeneratorImpl {
	return &ShortCodeGeneratorImpl{
		length: length,
	}
}

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func (s *ShortCodeGeneratorImpl) GenerateShortCode() string {
	length := len(chars)
	result := make([]byte, s.length)
	for i := 0; i < s.length; i++ {
		result[i] = chars[rand.Intn(length)]
	}
	return string(result)
}
