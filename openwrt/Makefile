include $(TOPDIR)/rules.mk

PKG_NAME:=zigbee-go-home
PKG_VERSION:=0.1.0
PKG_RELEASE:=1

PKG_SOURCE:=$(PKG_NAME)-$(PKG_VERSION).tar.gz
PKG_SOURCE_URL:=https://codeload.github.com/lmahmutov/zigbee-go-home/tar.gz/v$(PKG_VERSION)?
PKG_HASH:=skip

PKG_LICENSE:=MIT
PKG_LICENSE_FILES:=LICENSE
PKG_MAINTAINER:=lmahmutov <lenz1986@gmail.com>

PKG_BUILD_DEPENDS:=golang/host
PKG_BUILD_PARALLEL:=1
PKG_BUILD_FLAGS:=no-mips16

GO_PKG:=zigbee-go-home
GO_PKG_BUILD_PKG:=zigbee-go-home/cmd/zigbee-home
GO_PKG_LDFLAGS_X:=main.version=$(PKG_VERSION)

include $(INCLUDE_DIR)/package.mk
include $(TOPDIR)/feeds/packages/lang/golang/golang-package.mk

define Package/zigbee-go-home
  SECTION:=net
  CATEGORY:=Network
  TITLE:=Zigbee smart home coordinator
  URL:=https://github.com/lmahmutov/zigbee-go-home
  DEPENDS:=$(GO_ARCH_DEPENDS) +kmod-usb-acm
endef

define Package/zigbee-go-home/description
  Zigbee smart home coordinator using nRF52840 NCP over USB.
  Web dashboard, REST API, WebSocket events, MQTT bridge with
  Home Assistant autodiscovery.
endef

define Package/zigbee-go-home/conffiles
/etc/zigbee-go-home/config.yaml
endef

define Package/zigbee-go-home/install
	$(call GoPackage/Package/Install/Bin,$(1))

	$(INSTALL_DIR) $(1)/etc/zigbee-go-home
	$(INSTALL_CONF) $(CURDIR)/files/config.yaml $(1)/etc/zigbee-go-home/config.yaml

	$(INSTALL_DIR) $(1)/etc/init.d
	$(INSTALL_BIN) $(CURDIR)/files/zigbee-go-home.init $(1)/etc/init.d/zigbee-go-home

	$(INSTALL_DIR) $(1)/var/lib/zigbee-go-home
endef

$(eval $(call GoBinPackage,zigbee-go-home))
$(eval $(call BuildPackage,zigbee-go-home))
