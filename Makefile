PREFIX ?= $(HOME)/.lazado

.PHONY: install uninstall update help

install:
	@mkdir -p $(PREFIX)/lib $(PREFIX)/config
	@cp lazado.sh $(PREFIX)/
	@cp lib/*.sh $(PREFIX)/lib/
	@cp config/*.template $(PREFIX)/config/
	@echo ""
	@echo "Installed to $(PREFIX)"
	@echo ""
	@echo "Add to your ~/.bashrc:"
	@echo '  export LAZADO_DIR="$(PREFIX)"'
	@echo '  [ -s "$$LAZADO_DIR/lazado.sh" ] && \. "$$LAZADO_DIR/lazado.sh"'
	@echo ""
	@echo "Then run: source ~/.bashrc && ado-init"

uninstall:
	@rm -rf $(PREFIX)
	@echo "Removed $(PREFIX)"
	@echo "Remember to remove the lazado lines from your ~/.bashrc"

update:
	@if [ -d "$(PREFIX)/.git" ]; then \
		cd $(PREFIX) && git pull origin main; \
	else \
		$(MAKE) install; \
	fi
	@echo "Updated lazado."

help:
	@echo "lazado Makefile targets:"
	@echo "  make install    Install to $(PREFIX)"
	@echo "  make uninstall  Remove installation"
	@echo "  make update     Update to latest version"
