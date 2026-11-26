/**
 * SiteCraft Admin Blocks Editor
 * Interactive Visual Builder, Live Preview, and OKLCH Design Token Tuner
 */

// 1. OKLCH Color Space & Contrast Guard Utility
window.SiteOklchGuard = window.SiteOklchGuard || {
	srgbToLinear: function(c) {
		c = Math.max(0, Math.min(1, c));
		return (c <= 0.04045) ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4);
	},
	linearToSrgb: function(c) {
		c = Math.max(0, Math.min(1, c));
		return (c <= 0.0031308) ? 12.92 * c : 1.055 * Math.pow(c, 1 / 2.4) - 0.055;
	},
	hexToRgb: function(hex) {
		if (!hex || hex.length < 7) return { r: 0, g: 0, b: 0 };
		return {
			r: parseInt(hex.slice(1, 3), 16),
			g: parseInt(hex.slice(3, 5), 16),
			b: parseInt(hex.slice(5, 7), 16)
		};
	},
	rgbToHex: function(r, g, b) {
		var toHex = function(v) {
			var s = Math.max(0, Math.min(255, Math.round(v))).toString(16);
			return s.length === 1 ? '0' + s : s;
		};
		return '#' + toHex(r) + toHex(g) + toHex(b);
	},
	rgbToOklch: function(r, g, b) {
		var rL = this.srgbToLinear(r / 255);
		var gL = this.srgbToLinear(g / 255);
		var bL = this.srgbToLinear(b / 255);

		var l = 0.4122214708 * rL + 0.5363325363 * gL + 0.0514459929 * bL;
		var m = 0.2119034982 * rL + 0.6806995451 * gL + 0.1073969566 * bL;
		var s = 0.0883024619 * rL + 0.2817188376 * gL + 0.6299787005 * bL;

		var l_ = Math.cbrt(l);
		var m_ = Math.cbrt(m);
		var s_ = Math.cbrt(s);

		var L = 0.2104542553 * l_ + 0.7936177850 * m_ - 0.0040720468 * s_;
		var a = 1.9779984951 * l_ - 2.4285922050 * m_ + 0.4505937099 * s_;
		var bO = 0.0259040371 * l_ + 0.7827717662 * m_ - 0.8086757660 * s_;

		var C = Math.sqrt(a * a + bO * bO);
		var H = (Math.atan2(bO, a) * 180 / Math.PI + 360) % 360;

		return { l: Math.max(0, Math.min(1, L)), c: C, h: H };
	},
	oklchToRgb: function(L, C, H) {
		L = Math.max(0, Math.min(1, L));
		var toLinear = function(lVal, cVal, hVal) {
			var a = cVal * Math.cos(hVal * Math.PI / 180);
			var bO = cVal * Math.sin(hVal * Math.PI / 180);
			var l_ = lVal + 0.3963377774 * a + 0.2158037573 * bO;
			var m_ = lVal - 0.1055613458 * a - 0.0638541728 * bO;
			var s_ = lVal - 0.0894841775 * a - 1.2914855480 * bO;
			var l = l_ * l_ * l_;
			var m = m_ * m_ * m_;
			var s = s_ * s_ * s_;
			return {
				rL: +4.0767416621 * l - 3.3077115913 * m + 0.2309699292 * s,
				gL: -1.2684380046 * l + 2.6097574011 * m - 0.3413193965 * s,
				bL: -0.0041960863 * l - 0.7034186147 * m + 1.7076147010 * s
			};
		};
		var inGamut = function(t) {
			var eps = 0.0008;
			return t.rL >= -eps && t.rL <= (1 + eps) &&
			       t.gL >= -eps && t.gL <= (1 + eps) &&
			       t.bL >= -eps && t.bL <= (1 + eps);
		};
		var safeC = Math.max(0, C);
		var test = toLinear(L, safeC, H);
		if (!inGamut(test)) {
			var low = 0, high = safeC;
			for (var iter = 0; iter < 10; iter++) {
				var mid = (low + high) / 2;
				if (inGamut(toLinear(L, mid, H))) {
					low = mid;
				} else {
					high = mid;
				}
			}
			safeC = low;
			test = toLinear(L, safeC, H);
		}
		var r = Math.round(this.linearToSrgb(test.rL) * 255);
		var g = Math.round(this.linearToSrgb(test.gL) * 255);
		var b = Math.round(this.linearToSrgb(test.bL) * 255);
		var clampedRL = Math.max(0, Math.min(1, test.rL));
		var clampedGL = Math.max(0, Math.min(1, test.gL));
		var clampedBL = Math.max(0, Math.min(1, test.bL));
		var lum = 0.2126 * clampedRL + 0.7152 * clampedGL + 0.0722 * clampedBL;

		return {
			r: r, g: g, b: b,
			hex: this.rgbToHex(r, g, b),
			lum: Math.max(0, lum),
			chroma: safeC
		};
	},
	getOnColor: function(hex) {
		var lum = this.getLuminanceFromHex(hex);
		var ratioWithDark = this.calcRatio(lum, 0.009);
		var ratioWithWhite = this.calcRatio(lum, 1.0);
		return (ratioWithDark > ratioWithWhite) ? '#0f172a' : '#ffffff';
	},
	generateScale: function(H, C) {
		var steps = [
			{ name: '50', l: 0.98 },
			{ name: '100', l: 0.95 },
			{ name: '200', l: 0.90 },
			{ name: '300', l: 0.82 },
			{ name: '400', l: 0.70 },
			{ name: '500', l: 0.58 },
			{ name: '600', l: 0.48 },
			{ name: '700', l: 0.38 },
			{ name: '800', l: 0.28 },
			{ name: '900', l: 0.18 },
			{ name: '950', l: 0.10 }
		];
		var list = [];
		for (var i = 0; i < steps.length; i++) {
			var s = steps[i];
			var res = this.oklchToRgb(s.l, C, H);
			list.push({
				name: s.name,
				lPercent: Math.round(s.l * 100),
				hex: res.hex,
				textColor: s.l > 0.65 ? '#0f172a' : '#ffffff'
			});
		}
		return list;
	},
	getLuminanceFromHex: function(hex) {
		var rgb = this.hexToRgb(hex);
		var rL = this.srgbToLinear(rgb.r / 255);
		var gL = this.srgbToLinear(rgb.g / 255);
		var bL = this.srgbToLinear(rgb.b / 255);
		return 0.2126 * rL + 0.7152 * gL + 0.0722 * bL;
	},
	calcRatio: function(lum1, lum2) {
		var max = Math.max(lum1, lum2);
		var min = Math.min(lum1, lum2);
		return (max + 0.05) / (min + 0.05);
	},
	getSafeRange: function(H, C, bgHex, minAA, minAAA) {
		var bgLum = this.getLuminanceFromHex(bgHex);
		var isLightBg = bgLum > 0.5;
		var targetAA = minAA || 4.5;
		var targetAAA = minAAA || 7.0;

		var findLForRatio = (targetRatio) => {
			var low = 0.0, high = 1.0;
			for (var i = 0; i < 20; i++) {
				var mid = (low + high) / 2;
				var res = this.oklchToRgb(mid, C, H);
				var ratio = this.calcRatio(res.lum, bgLum);
				if (isLightBg) {
					if (ratio > targetRatio) low = mid;
					else high = mid;
				} else {
					if (ratio > targetRatio) high = mid;
					else low = mid;
				}
			}
			return (low + high) / 2;
		};

		var lAA = findLForRatio(targetAA);
		var lAAA = findLForRatio(targetAAA);

		return {
			isLightBg: isLightBg,
			aaPercent: Math.max(0, Math.min(100, Math.round(lAA * 100))),
			aaaPercent: Math.max(0, Math.min(100, Math.round(lAAA * 100)))
		};
	},
	getGradientCss: function(H, C) {
		var stops = [];
		for (var step = 5; step <= 95; step += 15) {
			var rgb = this.oklchToRgb(step / 100, C, H);
			stops.push(rgb.hex + ' ' + step + '%');
		}
		return 'linear-gradient(to right, #000000 0%, ' + stops.join(', ') + ', #ffffff 100%)';
	},
	snapToTarget: function(H, C, bgHex, level, minRatio) {
		var targetRatio = minRatio || (level === 'aaa' ? 7.0 : 4.5);
		var range = this.getSafeRange(H, C, bgHex, targetRatio, targetRatio);
		var targetPercent = range.aaPercent;
		if (range.isLightBg) {
			targetPercent = Math.max(5, targetPercent - 3);
		} else {
			targetPercent = Math.min(95, targetPercent + 3);
		}
		var L = targetPercent / 100;
		var rgb = this.oklchToRgb(L, C, H);
		return { l: L, hex: rgb.hex };
	}
};

// 2. Helper to open block editor modal from external contexts
window.openBlockEditor = function(id, title, content, bType, variant) {
	var idInp = document.getElementById('edit-block-id');
	if (idInp) idInp.value = id;
	var typeInp = document.getElementById('edit-block-type');
	if (typeInp) typeInp.value = bType;
	var titleInp = document.getElementById('edit-block-title');
	if (titleInp) titleInp.value = title;
	var contentInp = document.getElementById('edit-block-content');
	if (contentInp) contentInp.value = content;
	var variantRow = document.getElementById('edit-variant-row');
	var variantSelect = document.getElementById('edit-block-variant');
	if (variantRow && variantSelect) {
		if (bType === 'hero') {
			variantRow.style.display = 'block';
			variantSelect.value = variant || 'classic';
		} else {
			variantRow.style.display = 'none';
		}
	}
	var dlg = document.getElementById('edit-block-dialog');
	if (dlg) {
		var root = dlg.closest('[x-data]');
		if (root && window.Alpine) {
			try {
				var d = Alpine.$data(root);
				if (d) {
					d.editBlock = { id: id, title: title, content: content, type: bType, variant: variant || 'classic' };
				}
			} catch (e) {}
		}
		dlg.showModal();
	}
};

// 3. Alpine.data component registration
(function() {
	function registerBlocksEditor() {
		if (!window.Alpine || window._adminBlocksEditorRegistered) return;
		window._adminBlocksEditorRegistered = true;

		Alpine.data('adminBlocksEditor', function(config) {
			config = config || {};
			var theme = config.theme || 'main';
			var primary = config.primary || '#1e293b';
			var accent = config.accent || '#2563eb';
			var surface = config.surface || '#ffffff';
			var flavor = config.flavor || 'corporate';
			var font = config.font || 'inter';

			return {
				activeSubTab: 'blocks',
				canvasDevice: 'desktop',
				containerWidth: 900,
				selectedBlockId: '',
				responsiveView: 'both',
				editBlock: {
					id: '',
					title: '',
					content: '',
					type: '',
					variant: ''
				},
				themePreset: theme,
				flavorPreset: flavor,
				fontPreset: font,
				tokens: {
					primary: primary,
					accent: accent,
					surface: surface,
					radiusBtn: '8px',
					radiusCard: '24px'
				},
				contrastRatio: 7.2,
				contrastPass: true,
				activeTuner: '',
				inspectRole: 'primary',
				brandHue: 220,
				brandChroma: 0.16,
				brandLightness: 35,
				lightnessGradient: '',
				safeZones: {
					isLightBg: true,
					aaPercent: 55,
					aaaPercent: 40
				},
				radarStatus: {
					checked: false,
					hasOverflow: false,
					diff: 0,
					culprits: []
				},
				saveStatus: '',
				seedColor: accent,
				isGrayscale: false,
				harmonyPreset: 'brand',
				neutralTemp: 'brand',
				accentScale: [],
				copiedStep: '',
				showProTools: false,
				presetSwatches: {
					primary: [
						{ name: 'Slate', hex: '#0f172a' },
						{ name: 'Midnight', hex: '#09090b' },
						{ name: 'Navy', hex: '#0f1e36' },
						{ name: 'Emerald', hex: '#06231a' },
						{ name: 'Burgundy', hex: '#2c0d16' },
						{ name: 'Espresso', hex: '#231610' },
						{ name: 'Zinc', hex: '#18181b' },
						{ name: 'Indigo', hex: '#1e1b4b' }
					],
					accent: [
						{ name: 'Amber', hex: '#d97706' },
						{ name: 'Blue', hex: '#2563eb' },
						{ name: 'Emerald', hex: '#059669' },
						{ name: 'Crimson', hex: '#dc2626' },
						{ name: 'Indigo', hex: '#4f46e5' },
						{ name: 'Rose', hex: '#e11d48' },
						{ name: 'Cyan', hex: '#0891b2' },
						{ name: 'Dark', hex: '#0f172a' }
					],
					surface: [
						{ name: 'White', hex: '#ffffff' },
						{ name: 'Parchment', hex: '#faf8f5' },
						{ name: 'Ice Slate', hex: '#f8fafc' },
						{ name: 'Neutral', hex: '#f9fafb' },
						{ name: 'Night', hex: '#0b0f19' },
						{ name: 'Dark Zinc', hex: '#18181b' }
					]
				},
				imageExtractor: {
					isOpen: false,
					isProcessing: false,
					isDragOver: false,
					imageSrc: '',
					fileName: '',
					dominantColors: [],
					curatedPalettes: [],
					selectedPaletteId: '',
					sampleRole: 'accent',
					pickedPixelHex: '',
					statusMsg: '',
					errorMsg: ''
				},
				init() {
					this.$nextTick(() => { this.updateContainerWidth(); });
					window.addEventListener('resize', () => { this.updateContainerWidth(); });
					window.addEventListener('message', (e) => {
						if (e.data && e.data.type === 'PREVIEW_BLOCK_CLICKED') {
							if (e.data.blockId) {
								this.selectBlock(e.data.blockId, e.data.blockType, false);
							}
						}
						if (e.data && e.data.type === 'SITE_OVERFLOW_REPORT') {
							this.radarStatus = {
								checked: true,
								hasOverflow: !!e.data.hasOverflow,
								diff: e.data.diff || 0,
								culprits: e.data.culprits || []
							};
						}
					});
					window.addEventListener('paste', (e) => {
						if (!this.imageExtractor.isOpen) return;
						if (e.clipboardData && e.clipboardData.items) {
							for (let i = 0; i < e.clipboardData.items.length; i++) {
								const item = e.clipboardData.items[i];
								if (item.type.indexOf('image') !== -1) {
									const file = item.getAsFile();
									if (file) {
										this.handleImageFile(file);
										break;
									}
								}
							}
						}
					});
					this.$nextTick(() => {
						this.syncFromHex(this.tokens[this.inspectRole] || this.tokens.primary);
						this.updateScale();
					});
				},
				updateContainerWidth() {
					const el = this.$refs.canvasContainer;
					if (el) {
						this.containerWidth = el.clientWidth || 800;
					}
				},
				get scaleFactor() {
					if (this.canvasDevice === 'desktop') {
						const targetWidth = 1280;
						const available = Math.max(300, (this.containerWidth || 800) - 8);
						return Math.min(1, Math.max(0.2, available / targetWidth));
					}
					return 1;
				},
				setDevice(dev) {
					this.canvasDevice = dev;
					this.$nextTick(() => {
						this.updateContainerWidth();
						setTimeout(() => { this.triggerRadarCheck(); }, 350);
					});
				},
				triggerRadarCheck() {
					const frame = document.getElementById('preview-canvas');
					if (frame && frame.contentWindow) {
						frame.contentWindow.postMessage({ type: 'SITE_CHECK_OVERFLOW' }, '*');
					}
				},
				selectBlock(id, bType, scrollIframe = true) {
					this.selectedBlockId = id;
					const card = document.getElementById('block-card-' + id);
					if (card) {
						card.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
					}
					if (scrollIframe) {
						const frame = document.getElementById('preview-canvas');
						if (frame && frame.contentWindow) {
							frame.contentWindow.postMessage({
								type: 'SITE_SCROLL_TO_BLOCK',
								blockId: id,
								blockType: bType
							}, '*');
						}
					}
				},
				openEdit(id, title, content, bType, variant) {
					this.editBlock = {
						id: id,
						title: title,
						content: content,
						type: bType,
						variant: variant
					};
					const dlg = document.getElementById('edit-block-dialog');
					if (dlg) dlg.showModal();
				},
				sendTokenUpdate(key, val) {
					const frame = document.getElementById('preview-canvas');
					if (frame && frame.contentWindow) {
						frame.contentWindow.postMessage({ type: 'SITE_TOKEN_UPDATE', key: key, value: val }, '*');
					}
					if (key === '--color-accent') {
						this.updateScale();
					}
					this.syncFromHex(this.tokens[this.inspectRole]);
				},
				sendThemeUpdate(newTheme) {
					this.themePreset = newTheme;
					const frame = document.getElementById('preview-canvas');
					if (frame && frame.contentWindow) {
						frame.contentWindow.postMessage({ type: 'SITE_THEME_UPDATE', theme: newTheme }, '*');
					}
				},
				applyPresetStyle(preset) {
					if (!preset) return;
					this.themePreset = preset.theme;
					this.flavorPreset = preset.flavor;
					this.fontPreset = preset.font;
					this.tokens.primary = preset.primary;
					this.tokens.accent = preset.accent;
					this.tokens.surface = preset.surface;

					const frame = document.getElementById('preview-canvas');
					if (frame && frame.contentWindow) {
						frame.contentWindow.postMessage({ type: 'SITE_THEME_UPDATE', theme: preset.theme }, '*');
						frame.contentWindow.postMessage({ type: 'SITE_FLAVOR_UPDATE', flavor: preset.flavor }, '*');
						frame.contentWindow.postMessage({ type: 'SITE_FONT_UPDATE', font: preset.font }, '*');
						frame.contentWindow.postMessage({ type: 'SITE_TOKEN_UPDATE', key: '--color-primary', value: preset.primary }, '*');
						frame.contentWindow.postMessage({ type: 'SITE_TOKEN_UPDATE', key: '--color-accent', value: preset.accent }, '*');
						frame.contentWindow.postMessage({ type: 'SITE_TOKEN_UPDATE', key: '--color-surface', value: preset.surface }, '*');
					}
					this.updateScale();
					this.syncFromHex(this.tokens[this.inspectRole]);

					// Persist immediately in 1 click
					var fd = new FormData();
					fd.append('preset_id', preset.id);
					this.saveStatus = 'Применение стиля...';
					fetch('/admin/api/theme/apply', { method: 'POST', body: fd })
						.then(r => r.text())
						.then(html => {
							this.saveStatus = 'Стиль «' + preset.name + '» применен ✓';
							const banner = document.getElementById('theme-rollback-banner');
							if (banner) {
								banner.innerHTML = html;
							}
							setTimeout(() => { this.saveStatus = ''; }, 4000);
						})
						.catch(() => {
							this.saveStatus = 'Ошибка применения';
						});
				},
				sendFlavorUpdate(flv) {
					this.flavorPreset = flv;
					const frame = document.getElementById('preview-canvas');
					if (frame && frame.contentWindow) {
						frame.contentWindow.postMessage({ type: 'SITE_FLAVOR_UPDATE', flavor: flv }, '*');
					}
				},
				sendFontUpdate(fnt) {
					this.fontPreset = fnt;
					const frame = document.getElementById('preview-canvas');
					if (frame && frame.contentWindow) {
						frame.contentWindow.postMessage({ type: 'SITE_FONT_UPDATE', font: fnt }, '*');
					}
				},
				toggleGrayscale() {
					this.isGrayscale = !this.isGrayscale;
					const frame = document.getElementById('preview-canvas');
					if (frame && frame.contentWindow) {
						frame.contentWindow.postMessage({ type: 'SITE_GRAYSCALE_UPDATE', enabled: this.isGrayscale }, '*');
					}
				},
				copyHex(hex, step) {
					if (navigator && navigator.clipboard) {
						navigator.clipboard.writeText(hex);
						this.copiedStep = step;
						setTimeout(() => { if (this.copiedStep === step) this.copiedStep = ''; }, 1600);
					}
				},
				updateScale() {
					if (!window.SiteOklchGuard) return;
					var rgb = window.SiteOklchGuard.hexToRgb(this.tokens.accent || '#2563eb');
					var oklch = window.SiteOklchGuard.rgbToOklch(rgb.r, rgb.g, rgb.b);
					this.accentScale = window.SiteOklchGuard.generateScale(oklch.h, oklch.c);
				},
				getOnAccentContrast() {
					if (!window.SiteOklchGuard) return { pass: true, ratio: 7.0, label: 'AAA', textColor: '#ffffff' };
					var onColor = window.SiteOklchGuard.getOnColor(this.tokens.accent);
					var lumAccent = window.SiteOklchGuard.getLuminanceFromHex(this.tokens.accent);
					var lumText = window.SiteOklchGuard.getLuminanceFromHex(onColor);
					var r = window.SiteOklchGuard.calcRatio(lumAccent, lumText);
					var rounded = Math.round(r * 10) / 10;
					return {
						ratio: rounded,
						pass: rounded >= 4.5,
						label: rounded >= 7.0 ? 'AAA' : (rounded >= 4.5 ? 'AA' : 'Fail'),
						textColor: onColor
					};
				},
				applyHarmony(mode) {
					this.harmonyPreset = mode;
					this.generate603010(this.tokens.accent);
				},
				setNeutralTemp(temp) {
					this.neutralTemp = temp;
					this.generate603010(this.tokens.accent);
				},
				getRoleContrast(role) {
					if (!window.SiteOklchGuard) return { ratio: 4.5, pass: true, label: 'AA' };
					var hex = this.tokens[role] || '#000000';
					var lum = window.SiteOklchGuard.getLuminanceFromHex(hex);
					var compareHex = role === 'surface' ? (this.tokens.primary || '#000000') : (this.tokens.surface || '#ffffff');
					var bgLum = window.SiteOklchGuard.getLuminanceFromHex(compareHex);
					var r = window.SiteOklchGuard.calcRatio(lum, bgLum);
					var rounded = Math.round(r * 10) / 10;
					var target = (role === 'accent') ? 3.0 : 4.5;
					var aaaTarget = (role === 'accent') ? 4.5 : 7.0;
					return {
						ratio: rounded,
						pass: rounded >= target,
						label: rounded >= aaaTarget ? 'AAA' : (rounded >= target ? 'AA' : 'Fail')
					};
				},
				applyRoleHex(role, hex) {
					if (!hex) return;
					hex = hex.trim();
					if (!hex.startsWith('#')) hex = '#' + hex;
					if (hex.length !== 7) return;
					this.tokens[role] = hex;
					var key = role === 'primary' ? '--color-primary' : (role === 'accent' ? '--color-accent' : '--color-surface');
					this.sendTokenUpdate(key, hex);
					if (this.inspectRole === role) {
						this.syncFromHex(hex);
					}
				},
				async pickEyeDropper(role) {
					if (window.EyeDropper) {
						try {
							const eyeDropper = new window.EyeDropper();
							const result = await eyeDropper.open();
							if (result && result.sRGBHex) {
								this.applyRoleHex(role, result.sRGBHex);
							}
						} catch (e) {}
					} else {
						const inp = document.getElementById('native-picker-' + role);
						if (inp) inp.click();
					}
				},
				toggleTuner(role) {
					if (this.activeTuner === role) {
						this.activeTuner = '';
					} else {
						this.activeTuner = role;
						this.inspectRole = role;
						this.syncFromHex(this.tokens[role]);
					}
				},
				syncFromHex(hex) {
					if (!window.SiteOklchGuard) return;
					var rgb = window.SiteOklchGuard.hexToRgb(hex);
					var oklch = window.SiteOklchGuard.rgbToOklch(rgb.r, rgb.g, rgb.b);
					this.brandHue = Math.round(oklch.h);
					this.brandChroma = Math.max(0.04, Math.min(0.24, Math.round(oklch.c * 100) / 100 || 0.15));
					this.brandLightness = Math.round(oklch.l * 100);
					this.recalculateOklch(false);
				},
				recalculateOklch(sendUpdate = false) {
					if (!window.SiteOklchGuard) return;
					var compareBg = this.inspectRole === 'surface' ? this.tokens.primary : this.tokens.surface;
					var isAccent = this.inspectRole === 'accent';
					this.safeZones = window.SiteOklchGuard.getSafeRange(this.brandHue, this.brandChroma, compareBg, isAccent ? 3.0 : 4.5, isAccent ? 4.5 : 7.0);
					this.lightnessGradient = window.SiteOklchGuard.getGradientCss(this.brandHue, this.brandChroma);
					var res = window.SiteOklchGuard.oklchToRgb(this.brandLightness / 100, this.brandChroma, this.brandHue);
					this.tokens[this.inspectRole] = res.hex;
					var bgLum = window.SiteOklchGuard.getLuminanceFromHex(compareBg);
					var ratio = window.SiteOklchGuard.calcRatio(res.lum, bgLum);
					this.contrastRatio = Math.round(ratio * 10) / 10;
					this.contrastPass = ratio >= (isAccent ? 3.0 : 4.5);
					if (sendUpdate) {
						var key = this.inspectRole === 'primary' ? '--color-primary' : (this.inspectRole === 'accent' ? '--color-accent' : '--color-surface');
						const frame = document.getElementById('preview-canvas');
						if (frame && frame.contentWindow) {
							frame.contentWindow.postMessage({ type: 'SITE_TOKEN_UPDATE', key: key, value: res.hex }, '*');
						}
						if (this.inspectRole === 'accent') {
							this.updateScale();
						}
					}
				},
				onHueChange(val) {
					this.brandHue = parseInt(val);
					this.recalculateOklch(true);
				},
				onLightnessChange(val) {
					this.brandLightness = parseInt(val);
					this.recalculateOklch(true);
				},
				snapTo(level) {
					if (!window.SiteOklchGuard) return;
					var compareBg = this.inspectRole === 'surface' ? this.tokens.primary : this.tokens.surface;
					var minRatio = (this.inspectRole === 'accent') ? (level === 'aaa' ? 4.5 : 3.0) : (level === 'aaa' ? 7.0 : 4.5);
					var res = window.SiteOklchGuard.snapToTarget(this.brandHue, this.brandChroma, compareBg, level, minRatio);
					if (res) {
						this.brandLightness = Math.round(res.l * 100);
						this.recalculateOklch(true);
					}
				},
				setInspectRole(role) {
					this.inspectRole = role;
					this.activeTuner = role;
					this.syncFromHex(this.tokens[role]);
				},
				autoFixRole(role) {
					this.inspectRole = role;
					this.activeTuner = role;
					this.syncFromHex(this.tokens[role]);
					this.snapTo('aa');
				},
				autoFixContrast() {
					this.snapTo('aa');
				},
				generate603010(seed) {
					if (!seed) return;
					this.tokens.accent = seed;
					if (window.SiteOklchGuard) {
						var rgb = window.SiteOklchGuard.hexToRgb(seed);
						var oklch = window.SiteOklchGuard.rgbToOklch(rgb.r, rgb.g, rgb.b);
						var h = oklch.h;
						var isDarkTheme = this.themePreset === 'main-night';

						// 1. Harmonies are only a start (угол hue для текста)
						var textHue = h;
						if (this.harmonyPreset === 'complementary') textHue = (h + 180) % 360;
						else if (this.harmonyPreset === 'analogous') textHue = (h + 30) % 360;
						else if (this.harmonyPreset === 'split') textHue = (h + 150) % 360;
						else if (this.harmonyPreset === 'triadic') textHue = (h + 120) % 360;

						// 2. Neutrals with a temperature (угол и примесь для фона)
						var bgHue = h;
						var bgChroma = 0.008;
						var textChroma = 0.035; // One Loud Color: умеренная насыщенность текста
						if (this.neutralTemp === 'clean') {
							bgChroma = 0.0;
							textChroma = 0.0;
						} else if (this.neutralTemp === 'warm') {
							bgHue = 65;
							bgChroma = 0.012;
						} else if (this.neutralTemp === 'cool') {
							bgHue = 220;
							bgChroma = 0.012;
						}

						if (isDarkTheme) {
							// Темная тема: глубокий брендовый фон и светлый читаемый текст
							var bgRes = window.SiteOklchGuard.oklchToRgb(0.12, Math.max(0.01, bgChroma * 2), bgHue);
							var primRes = window.SiteOklchGuard.oklchToRgb(0.96, textChroma, textHue);
							this.tokens.surface = bgRes.hex;
							this.tokens.primary = primRes.hex;
						} else {
							// Светлая тема: чистый/тонированный светлый фон и глубокий читаемый текст
							var bgRes = window.SiteOklchGuard.oklchToRgb(0.985, bgChroma, bgHue);
							// Светлота 18% с тонким оттенком дает статус AAA (≥7:1)
							var primRes = window.SiteOklchGuard.oklchToRgb(0.18, textChroma, textHue);
							this.tokens.surface = bgRes.hex;
							this.tokens.primary = primRes.hex;
						}
					} else {
						this.tokens.surface = '#ffffff';
						this.tokens.primary = '#0f172a';
					}
					this.sendTokenUpdate('--color-accent', this.tokens.accent);
					this.sendTokenUpdate('--color-primary', this.tokens.primary);
					this.sendTokenUpdate('--color-surface', this.tokens.surface);
					this.updateScale();
					this.syncFromHex(this.tokens[this.inspectRole]);
				},
				saveThemeTokens() {
					var fd = new FormData();
					fd.append('theme', this.themePreset);
					fd.append('primary_color', this.tokens.primary);
					fd.append('accent_color', this.tokens.accent);
					fd.append('design_flavor', this.flavorPreset);
					fd.append('font_family', this.fontPreset);
					fd.append('background_color', this.tokens.surface);
					this.saveStatus = 'Сохранение...';
					fetch('/admin/theme/tokens', { method: 'POST', body: fd })
						.then(r => r.text())
						.then(() => {
							this.saveStatus = 'Сохранено в БД ✓';
							setTimeout(() => { this.saveStatus = ''; }, 3500);
						})
						.catch(() => {
							this.saveStatus = 'Ошибка сохранения';
						});
				},
				toggleImageExtractor() {
					this.imageExtractor.isOpen = !this.imageExtractor.isOpen;
				},
				handleImageFile(file) {
					if (!file) return;
					if (!file.type || !file.type.startsWith('image/')) {
						this.imageExtractor.errorMsg = 'Пожалуйста, выберите файл изображения (PNG, JPG, WebP, SVG, GIF)';
						return;
					}
					this.imageExtractor.isProcessing = true;
					this.imageExtractor.errorMsg = '';
					this.imageExtractor.statusMsg = 'Анализ палитры изображения...';
					this.imageExtractor.fileName = file.name;

					const reader = new FileReader();
					reader.onload = (e) => {
						const dataUrl = e.target.result;
						this.imageExtractor.imageSrc = dataUrl;
						const img = new Image();
						img.onload = () => {
							this.extractPaletteFromImg(img, file);
						};
						img.src = dataUrl;
					};
					reader.readAsDataURL(file);
				},
				extractPaletteFromImg(img, rawFile) {
					try {
						// 1. Offscreen canvas extraction
						const maxDim = 120;
						let w = img.naturalWidth || img.width || 100;
						let h = img.naturalHeight || img.height || 100;
						if (w > maxDim || h > maxDim) {
							if (w > h) {
								h = Math.round((h * maxDim) / w);
								w = maxDim;
							} else {
								w = Math.round((w * maxDim) / h);
								h = maxDim;
							}
						}
						const canvas = document.createElement('canvas');
						canvas.width = Math.max(1, w);
						canvas.height = Math.max(1, h);
						const ctx = canvas.getContext('2d', { willReadFrequently: true });
						ctx.drawImage(img, 0, 0, w, h);
						const data = ctx.getImageData(0, 0, w, h).data;

						// 2. 5-bit color bucketing
						const buckets = {};
						let totalSampled = 0;
						for (let i = 0; i < data.length; i += 4) {
							const a = data[i + 3];
							if (a < 128) continue; // skip transparent
							const r = data[i];
							const g = data[i + 1];
							const b = data[i + 2];
							const key = ((r >> 3) << 10) | ((g >> 3) << 5) | (b >> 3);
							if (!buckets[key]) {
								buckets[key] = { rSum: 0, gSum: 0, bSum: 0, count: 0 };
							}
							buckets[key].rSum += r;
							buckets[key].gSum += g;
							buckets[key].bSum += b;
							buckets[key].count++;
							totalSampled++;
						}

						if (totalSampled === 0) {
							this.imageExtractor.errorMsg = 'Изображение полностью прозрачное';
							this.imageExtractor.isProcessing = false;
							return;
						}

						// 3. Cluster averages
						const clusters = [];
						for (const k in buckets) {
							const b = buckets[k];
							clusters.push({
								r: Math.round(b.rSum / b.count),
								g: Math.round(b.gSum / b.count),
								b: Math.round(b.bSum / b.count),
								count: b.count
							});
						}
						clusters.sort((a, b) => b.count - a.count);

						// 4. Merge visually close clusters
						const merged = [];
						const threshold = 28 * 28;
						for (const c of clusters) {
							let dup = false;
							for (const m of merged) {
								const dr = c.r - m.r;
								const dg = c.g - m.g;
								const db = c.b - m.b;
								if (dr * dr + dg * dg + db * db < threshold) {
									m.count += c.count;
									dup = true;
									break;
								}
							}
							if (!dup) {
								merged.push(c);
								if (merged.length >= 16) break;
							}
						}
						merged.sort((a, b) => b.count - a.count);

						// 5. Convert to color objects with metrics
						const dominant = [];
						for (const m of merged.slice(0, 8)) {
							const hex = window.SiteOklchGuard.rgbToHex(m.r, m.g, m.b);
							const lum = window.SiteOklchGuard.getLuminanceFromHex(hex);
							const rgb = window.SiteOklchGuard.hexToRgb(hex);
							const oklch = window.SiteOklchGuard.rgbToOklch(rgb.r, rgb.g, rgb.b);
							const pct = Math.round((m.count / totalSampled) * 1000) / 10;
							dominant.push({
								hex: hex,
								r: m.r,
								g: m.g,
								b: m.b,
								lum: lum,
								chroma: oklch.c,
								hue: Math.round(oklch.h),
								percent: pct,
								isDark: lum < 0.22
							});
						}

						this.imageExtractor.dominantColors = dominant;

						// 6. Identify Accent, Primary, Surface
						let bestAccent = dominant[0] ? dominant[0].hex : '#2563eb';
						let bestAccentScore = -1;
						for (const c of dominant) {
							if (c.lum > 0.03 && c.lum < 0.88) {
								const score = c.chroma * 4.0 + Math.min(1.0, c.percent / 15.0);
								if (score > bestAccentScore) {
									bestAccentScore = score;
									bestAccent = c.hex;
								}
							}
						}

						// Primary: deep dark color or tinted dark
						let bestPrimary = '#0f172a';
						let lowestLum = 1.0;
						for (const c of dominant) {
							if (c.hex !== bestAccent && c.lum < 0.15 && c.lum < lowestLum) {
								lowestLum = c.lum;
								bestPrimary = c.hex;
							}
						}
						if (lowestLum > 0.15) {
							const accRgb = window.SiteOklchGuard.hexToRgb(bestAccent);
							const accOklch = window.SiteOklchGuard.rgbToOklch(accRgb.r, accRgb.g, accRgb.b);
							const tinted = window.SiteOklchGuard.oklchToRgb(0.18, 0.035, accOklch.h);
							bestPrimary = tinted.hex;
						}

						// Surface: very light neutral tone or white
						let bestSurface = '#ffffff';
						for (const c of dominant) {
							if (c.lum > 0.88 && c.chroma < 0.04) {
								bestSurface = c.hex;
								break;
							}
						}

						// Curated palettes
						const accRgb = window.SiteOklchGuard.hexToRgb(bestAccent);
						const accOklch = window.SiteOklchGuard.rgbToOklch(accRgb.r, accRgb.g, accRgb.b);
						const accentHue = accOklch.h;

						const darkSurface = window.SiteOklchGuard.oklchToRgb(0.12, 0.02, accentHue).hex;
						const darkText = window.SiteOklchGuard.oklchToRgb(0.96, 0.02, accentHue).hex;
						const darkAccent = window.SiteOklchGuard.oklchToRgb(0.68, Math.max(0.16, accOklch.c), accentHue).hex;

						const warmSurface = window.SiteOklchGuard.oklchToRgb(0.985, 0.015, 65).hex; // Warm ivory
						const warmText = window.SiteOklchGuard.oklchToRgb(0.18, 0.025, 45).hex;

						this.imageExtractor.curatedPalettes = [
							{
								id: 'balanced',
								name: 'Сбалансированная (60-30-10)',
								desc: 'Чистый фон, четкий читаемый текст и выразительный акцент',
								accent: bestAccent,
								primary: bestPrimary,
								surface: bestSurface,
								theme: 'coral',
								flavor: 'soft'
							},
							{
								id: 'vibrant',
								name: 'Яркая / Брендовая',
								desc: 'Максимальная энергия цветов для активных проектов',
								accent: bestAccent,
								primary: bestPrimary,
								surface: '#ffffff',
								theme: 'coral',
								flavor: 'sharp'
							},
							{
								id: 'editorial',
								name: 'Мягкая / Редакционная',
								desc: 'Теплый благородный фон и сдержанный контраст',
								accent: bestAccent,
								primary: warmText,
								surface: warmSurface,
								theme: 'beauty',
								flavor: 'corporate'
							},
							{
								id: 'dark',
								name: 'Глубокая ночная (Dark)',
								desc: 'Премиальный темный фон со светящимся неоновым акцентом',
								accent: darkAccent,
								primary: darkText,
								surface: darkSurface,
								theme: 'main-night',
								flavor: 'soft'
							}
						];

						// Apply first palette to preview immediately
						this.applyCuratedPalette(this.imageExtractor.curatedPalettes[0]);
						this.imageExtractor.statusMsg = 'Палитра успешно извлечена!';
						this.imageExtractor.isProcessing = false;

						// In background: also call backend API if rawFile is present to verify
						if (rawFile) {
							const fd = new FormData();
							fd.append('image', rawFile);
							fetch('/admin/api/theme/extract-palette', { method: 'POST', body: fd })
								.then(r => r.json())
								.then(data => {
									if (data && data.success && data.palettes && data.palettes.length) {
										this.imageExtractor.curatedPalettes = data.palettes;
									}
								})
								.catch(() => {});
						}
					} catch (err) {
						this.imageExtractor.errorMsg = 'Ошибка обработки изображения: ' + (err.message || err);
						this.imageExtractor.isProcessing = false;
					}
				},
				applyCuratedPalette(pal) {
					if (!pal) return;
					this.imageExtractor.selectedPaletteId = pal.id;
					this.tokens.accent = pal.accent;
					this.tokens.primary = pal.primary;
					this.tokens.surface = pal.surface;
					if (pal.theme) this.themePreset = pal.theme;
					if (pal.flavor || pal.design_flavor) this.flavorPreset = pal.flavor || pal.design_flavor;

					const frame = document.getElementById('preview-canvas');
					if (frame && frame.contentWindow) {
						frame.contentWindow.postMessage({ type: 'SITE_THEME_UPDATE', theme: this.themePreset }, '*');
						frame.contentWindow.postMessage({ type: 'SITE_FLAVOR_UPDATE', flavor: this.flavorPreset }, '*');
						frame.contentWindow.postMessage({ type: 'SITE_TOKEN_UPDATE', key: '--color-primary', value: this.tokens.primary }, '*');
						frame.contentWindow.postMessage({ type: 'SITE_TOKEN_UPDATE', key: '--color-accent', value: this.tokens.accent }, '*');
						frame.contentWindow.postMessage({ type: 'SITE_TOKEN_UPDATE', key: '--color-surface', value: this.tokens.surface }, '*');
					}
					this.updateScale();
					this.syncFromHex(this.tokens[this.inspectRole]);
					this.imageExtractor.statusMsg = 'Палитра «' + pal.name + '» применена в предпросмотре!';
					setTimeout(() => { if (this.imageExtractor.statusMsg.startsWith('Палитра «')) this.imageExtractor.statusMsg = ''; }, 3500);
				},
				applyExtractedRole(hex, role) {
					role = role || this.imageExtractor.sampleRole || 'accent';
					this.applyRoleHex(role, hex);
					this.imageExtractor.statusMsg = 'Цвет ' + hex + ' назначен для ' + (role === 'accent' ? 'Акцента (10%)' : (role === 'primary' ? 'Текста (30%)' : 'Фона (60%)'));
					setTimeout(() => { this.imageExtractor.statusMsg = ''; }, 3000);
				},
				sampleImagePixel(event) {
					const img = event.target;
					const rect = img.getBoundingClientRect();
					const x = Math.floor((event.clientX - rect.left) * (img.naturalWidth / rect.width));
					const y = Math.floor((event.clientY - rect.top) * (img.naturalHeight / rect.height));

					const canvas = document.createElement('canvas');
					canvas.width = 1;
					canvas.height = 1;
					const ctx = canvas.getContext('2d', { willReadFrequently: true });
					ctx.drawImage(img, x, y, 1, 1, 0, 0, 1, 1);
					const pixel = ctx.getImageData(0, 0, 1, 1).data;
					const hex = window.SiteOklchGuard.rgbToHex(pixel[0], pixel[1], pixel[2]);

					this.imageExtractor.pickedPixelHex = hex;
					this.applyExtractedRole(hex, this.imageExtractor.sampleRole);
				},
				clearImageExtractor() {
					this.imageExtractor.imageSrc = '';
					this.imageExtractor.fileName = '';
					this.imageExtractor.dominantColors = [];
					this.imageExtractor.curatedPalettes = [];
					this.imageExtractor.selectedPaletteId = '';
					this.imageExtractor.statusMsg = '';
					this.imageExtractor.errorMsg = '';
					this.imageExtractor.pickedPixelHex = '';
				},
				loadSamplePaletteImage(type) {
					let svg = '';
					if (type === 'logo') {
						svg = '<svg xmlns="http://www.w3.org/2000/svg" width="400" height="240" viewBox="0 0 400 240"><rect width="100%" height="100%" fill="#ffffff"/><circle cx="120" cy="120" r="70" fill="#06b6d4"/><circle cx="200" cy="120" r="50" fill="#0f172a"/><circle cx="270" cy="120" r="60" fill="#3b82f6"/><text x="200" y="215" font-family="sans-serif" font-size="20" font-weight="bold" text-anchor="middle" fill="#0f172a">TECH LOGO</text></svg>';
					} else if (type === 'interior') {
						svg = '<svg xmlns="http://www.w3.org/2000/svg" width="400" height="240" viewBox="0 0 400 240"><rect width="100%" height="100%" fill="#fdfbf7"/><rect x="40" y="40" width="160" height="160" rx="16" fill="#b45309"/><rect x="220" y="40" width="140" height="100" rx="8" fill="#1c1917"/><circle cx="290" cy="175" r="25" fill="#d97706"/><text x="200" y="225" font-family="serif" font-size="16" text-anchor="middle" fill="#1c1917">LUXURY EDITORIAL</text></svg>';
					} else if (type === 'auto') {
						svg = '<svg xmlns="http://www.w3.org/2000/svg" width="400" height="240" viewBox="0 0 400 240"><rect width="100%" height="100%" fill="#18181b"/><polygon points="40,200 180,40 260,40 120,200" fill="#ea580c"/><rect x="240" y="100" width="120" height="60" fill="#ffffff"/><circle cx="100" cy="160" r="25" fill="#f97316"/><text x="200" y="225" font-family="sans-serif" font-size="18" font-weight="900" text-anchor="middle" fill="#ffffff">DYNAMIC AUTO</text></svg>';
					} else {
						svg = '<svg xmlns="http://www.w3.org/2000/svg" width="400" height="240" viewBox="0 0 400 240"><rect width="100%" height="100%" fill="#0a0f1d"/><circle cx="140" cy="110" r="65" fill="#e11d48"/><circle cx="240" cy="110" r="55" fill="#06b6d4"/><path d="M 60 190 Q 200 90 340 190" stroke="#38bdf8" stroke-width="8" fill="none"/><text x="200" y="220" font-family="sans-serif" font-size="18" font-weight="bold" text-anchor="middle" fill="#38bdf8">CYBERPUNK</text></svg>';
					}
					const dataUrl = 'data:image/svg+xml;charset=utf-8,' + encodeURIComponent(svg);
					this.imageExtractor.imageSrc = dataUrl;
					this.imageExtractor.fileName = 'sample_' + type + '.svg';
					this.imageExtractor.isProcessing = true;
					const img = new Image();
					img.onload = () => {
						this.extractPaletteFromImg(img, null);
					};
					img.src = dataUrl;
				},
				refreshCanvas() {
					const frame = document.getElementById('preview-canvas');
					if (frame) {
						frame.src = frame.src;
					}
				}
			};
		});
	}

	document.addEventListener('alpine:init', registerBlocksEditor);
	registerBlocksEditor();
})();
