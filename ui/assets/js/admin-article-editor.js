/**
 * SiteCraft Admin Article Editor
 * Transliteration, Cover Drag-and-Drop, Tag Management, and Quill.js Integration
 */

(function() {
	// 1. Cyrillic transliteration map for SEO-friendly URLs (ЧПУ)
	const ruTranslitMap = {
		'а': 'a', 'б': 'b', 'в': 'v', 'г': 'g', 'д': 'd',
		'е': 'e', 'ё': 'yo', 'ж': 'zh', 'з': 'z', 'и': 'i',
		'й': 'y', 'к': 'k', 'л': 'l', 'м': 'm', 'н': 'n',
		'о': 'o', 'п': 'p', 'р': 'r', 'с': 's', 'т': 't',
		'у': 'u', 'ф': 'f', 'х': 'kh', 'ц': 'ts', 'ч': 'ch',
		'ш': 'sh', 'щ': 'shch', 'ъ': '', 'ы': 'y', 'ь': '',
		'э': 'e', 'ю': 'yu', 'я': 'ya'
	};

	function transliterate(str) {
		if (!str) return '';
		let res = '';
		const lower = str.toLowerCase();
		for (let i = 0; i < lower.length; i++) {
			const ch = lower[i];
			if (ruTranslitMap[ch] !== undefined) {
				res += ruTranslitMap[ch];
			} else {
				res += ch;
			}
		}
		return res
			.replace(/[^a-z0-9\s-]/g, '')
			.trim()
			.replace(/[\s_]+/g, '-')
			.replace(/-+/g, '-');
	}

	window.transliterateCyrillic = transliterate;

	// 2. Article Editor Controller Factory
	function createArticleEditor(config) {
		config = config || {};
		return {
			title: config.initialTitle || '',
			slug: config.initialSlug || '',
			slugTouched: !config.isNew && config.initialSlug !== '',
			tags: Array.isArray(config.initialTags) ? config.initialTags : [],
			newTag: '',
			coverPreview: config.initialCover || '',
			removeCover: false,
			isDragging: false,
			showExtraSettings: false,
			quill: null,

			init() {
				this.initQuill();
			},

			transliterate(str) {
				return transliterate(str);
			},

			onTitleInput() {
				if (!this.slugTouched) {
					this.slug = this.transliterate(this.title);
				}
			},

			onSlugInput() {
				this.slugTouched = true;
			},

			regenerateSlug() {
				this.slug = this.transliterate(this.title);
				this.slugTouched = false;
			},

			addTag() {
				const val = this.newTag.trim().replace(/^#/, '').trim();
				if (val && !this.tags.includes(val)) {
					this.tags.push(val);
				}
				this.newTag = '';
			},

			removeTag(idx) {
				this.tags.splice(idx, 1);
			},

			suggestTag(name) {
				if (!this.tags.includes(name)) {
					this.tags.push(name);
				}
			},

			handleFileSelect(e) {
				const file = e.target.files && e.target.files[0];
				if (file) this.previewImage(file);
			},

			handleDrop(e) {
				this.isDragging = false;
				const file = e.dataTransfer && e.dataTransfer.files && e.dataTransfer.files[0];
				if (file) {
					if (this.$refs.coverFileInput) {
						this.$refs.coverFileInput.files = e.dataTransfer.files;
					}
					this.previewImage(file);
				}
			},

			previewImage(file) {
				if (!file || !file.type.startsWith('image/')) return;
				this.removeCover = false;
				const reader = new FileReader();
				reader.onload = (e) => {
					this.coverPreview = e.target.result;
				};
				reader.readAsDataURL(file);
			},

			clearCover() {
				this.coverPreview = '';
				this.removeCover = true;
				if (this.$refs.coverFileInput) {
					this.$refs.coverFileInput.value = '';
				}
			},

			initQuill() {
				const container = this.$refs.quillEditor;
				if (!container) return;
				if (container.__quill) return;

				if (typeof Quill === 'undefined') {
					if (!document.getElementById('quill-script')) {
						const s = document.createElement('script');
						s.id = 'quill-script';
						s.src = '/assets/js/quill.min.js';
						s.onload = () => this.initQuill();
						document.head.appendChild(s);
					}
					if (!document.getElementById('quill-style')) {
						const l = document.createElement('link');
						l.id = 'quill-style';
						l.rel = 'stylesheet';
						l.href = '/assets/css/quill.snow.css';
						document.head.appendChild(l);
					}
					setTimeout(() => this.initQuill(), 100);
					return;
				}

				const quill = new Quill(container, {
					theme: 'snow',
					placeholder: 'Напишите текст публикации: форматируйте абзацы, выделяйте заголовки H2/H3, списки, добавляйте изображения...',
					modules: {
						toolbar: [
							[{ 'header': [2, 3, false] }],
							['bold', 'italic', 'underline', 'strike'],
							[{ 'list': 'ordered'}, { 'list': 'bullet' }],
							['blockquote', 'link', 'image'],
							['clean']
						]
					}
				});

				container.__quill = quill;
				this.quill = quill;

				const initialHTML = this.$refs.contentInput ? (this.$refs.contentInput.value || '') : '';
				if (initialHTML) {
					quill.root.innerHTML = initialHTML;
				}

				quill.on('text-change', () => {
					const html = quill.root.innerHTML;
					if (this.$refs.contentInput) {
						this.$refs.contentInput.value = (html === '<p><br></p>') ? '' : html;
					}
				});
			}
		};
	}

	// Make available globally for backward compatibility
	window.articleEditor = createArticleEditor;

	// Register with Alpine
	function registerArticleEditor() {
		if (window.Alpine && !window._adminArticleEditorRegistered) {
			window._adminArticleEditorRegistered = true;
			Alpine.data('articleEditor', createArticleEditor);
		}
	}

	document.addEventListener('alpine:init', registerArticleEditor);
	registerArticleEditor();
})();
