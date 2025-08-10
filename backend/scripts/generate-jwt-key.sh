#!/bin/bash

echo "Generating secure JWT key..."

# Generate a 512-bit (64-byte) random key
JWT_KEY=$(openssl rand -base64 64)

echo "Generated JWT Secret Key:"
echo "$JWT_KEY"
echo ""
echo "To use this key:"
echo "1. Set environment variable:"
echo "   export JWT_SECRET_KEY=\"$JWT_KEY\""
echo ""
echo "2. Or add to your .env file:"
echo "   JWT_SECRET_KEY=$JWT_KEY"
echo ""
echo "3. Or set in docker-compose:"
echo "   JWT_SECRET_KEY: $JWT_KEY"
echo ""
echo "⚠️  IMPORTANT: Keep this key secret and secure!"
echo "⚠️  IMPORTANT: Use different keys for different environments!"
echo "⚠️  IMPORTANT: Rotate keys regularly in production!" 