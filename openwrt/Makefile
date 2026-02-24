include $(TOPDIR)/rules.mk

PKG_NAME:=zigbee-go-home
PKG_VERSION:=0.1.0
PKG_RELEASE:=1

PKG_BUILD_DIR:=$(BUILD_DIR)/$(PKG_NAME)-$(PKG_VERSION)

include $(INCLUDE_DIR)/package.mk

define Package/zigbee-go-home
  SECTION:=net
  CATEGORY:=Network
  TITLE:=Zigbee coordinator (Go)
  DEPENDS:=+kmod-usb-acm
  URL:=https://github.com/lmahmutov/zigbee-go-home
endef

define Package/zigbee-go-home/description
  Zigbee smart home coordinator using nRF52840 NCP over USB.
  Web dashboard, REST API, WebSocket events, MQTT bridge.
endef

GO_ARCH:=$(ARCH)
GO_ARM:=
GO_MIPS:=

ifeq ($(ARCH),aarch64)
  GO_ARCH:=arm64
endif
ifeq ($(ARCH),arm)
  GO_ARCH:=arm
  GO_ARM:=7
endif
ifeq ($(ARCH),mipsel)
  GO_ARCH:=mipsle
  GO_MIPS:=softfloat
endif
ifeq ($(ARCH),mips)
  GO_ARCH:=mips
  GO_MIPS:=softfloat
endif
ifeq ($(ARCH),x86_64)
  GO_ARCH:=amd64
endif

define Build/Compile
	cd $(PKG_BUILD_DIR) && \
	GOOS=linux GOARCH=$(GO_ARCH) GOARM=$(GO_ARM) GOMIPS=$(GO_MIPS) CGO_ENABLED=0 \
	go build -trimpath \
		-ldflags="-s -w -X main.version=$(PKG_VERSION)" \
		-o $(PKG_BUILD_DIR)/zigbee-home \
		./cmd/zigbee-home
endef

define Package/zigbee-go-home/install
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/zigbee-home $(1)/usr/bin/zigbee-home

	$(INSTALL_DIR) $(1)/etc/config
	$(INSTALL_CONF) ./files/config.yaml $(1)/etc/zigbee-go-home/config.yaml

	$(INSTALL_DIR) $(1)/etc/init.d
	$(INSTALL_BIN) ./files/zigbee-go-home.init $(1)/etc/init.d/zigbee-go-home

	$(INSTALL_DIR) $(1)/var/lib/zigbee-go-home
endef

$(eval $(call BuildPackage,zigbee-go-home))
