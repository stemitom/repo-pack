#!/bin/sh

# Set the repository and binary name
REPO="stemitom/repo-pack"
BINARY_NAME="repo-pack"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Check if the binary exists and delete it
if [ -f "/usr/local/bin/${BINARY_NAME}" ]; then
    echo -e "${YELLOW}Removing existing ${BINARY_NAME} binary...${NC}"
    sudo rm "/usr/local/bin/${BINARY_NAME}"
fi

echo "${BLUE}Getting the latest release information...${NC}"
LATEST_VERSION=$(curl -s https://api.github.com/repos/${REPO}/releases/latest | grep -o '"tag_name": ".*"' | awk -F'"' '{print $4}')

echo "${BLUE}Getting the machine architecture...${NC}"
ARCH=$(uname -m)
KERNEL=$(uname -s)

echo "${GREEN}Downloading the binary...${NC}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_VERSION}/${BINARY_NAME}-${KERNEL}-${ARCH}.tar.gz"
curl -# -L -o "${BINARY_NAME}.tar.gz" "${DOWNLOAD_URL}"

echo "${GREEN}Extracting the binary...${NC}"
tar -xzf "${BINARY_NAME}.tar.gz"

echo "${GREEN}Making the binary executable...${NC}"
chmod +x "${BINARY_NAME}"

printf "${YELLOW}Install the binary? (y/n) ${NC}"
read -r INSTALL </dev/tty

if [ "$INSTALL" = "y" ]; then
    echo "${GREEN}Installing the binary...${NC}"
    sudo mv "${BINARY_NAME}" /usr/local/bin/
else
    echo "${RED}Skipping installation...${NC}"
fi

echo "${GREEN}Cleaning up...${NC}"
rm "${BINARY_NAME}.tar.gz"

echo "${GREEN}Done!${NC}"
