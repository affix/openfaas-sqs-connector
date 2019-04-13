package sqs

import(
    "encoding/base64"
)

/**
 * Encodes the value in base64
 */
func Encode(value []byte) []byte {
    var length int = len(value)
    encoded := make([]byte, base64.URLEncoding.EncodedLen(length))
    base64.URLEncoding.Encode(encoded, value)
    return encoded
}


/**
 * Decodes the value using base64
 */
func Decode(value []byte) ([]byte, error) {
    var length int = len(value)
    decoded := make([]byte, base64.URLEncoding.DecodedLen(length))

    n, err := base64.URLEncoding.Decode(decoded, value)
    if err != nil {
        return nil, err
    }
    return decoded[:n], nil
}
