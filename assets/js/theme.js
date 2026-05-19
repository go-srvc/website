// Theme picker: reads/writes localStorage and sets data-theme on <html>.
// The selected attribute is applied pre-paint by a tiny inline script in the
// document head to avoid a flash of the default theme.
(() => {
	const KEY = 'go-srvc-theme';
	const select = document.getElementById('theme-select');
	if (!select) return;

	const stored = (() => {
		try { return localStorage.getItem(KEY); } catch (e) { return null; }
	})();
	select.value = stored || 'dracula';

	select.addEventListener('change', () => {
		const theme = select.value;
		document.documentElement.setAttribute('data-theme', theme);
		try { localStorage.setItem(KEY, theme); } catch (e) {}
	});
})();
