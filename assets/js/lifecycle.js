// Sequential playback of the lifecycle scenario animations. By default rows
// advance every 10 seconds. Hovering a row pins playback to that row until the
// pointer leaves.
(() => {
	const list = document.querySelector('.anim-list');
	if (!list) return;
	const figures = Array.from(list.querySelectorAll('.anim'));
	if (!figures.length) return;

	const cycleMs = 10000;
	let current = 0;
	let timer = null;

	const setActive = (i) => {
		current = ((i % figures.length) + figures.length) % figures.length;
		figures.forEach((el, idx) => el.classList.toggle('active', idx === current));
	};

	const start = () => {
		clearInterval(timer);
		timer = setInterval(() => setActive(current + 1), cycleMs);
	};

	const stop = () => clearInterval(timer);

	figures.forEach((fig, idx) => {
		fig.addEventListener('mouseenter', () => {
			stop();
			setActive(idx);
		});
		fig.addEventListener('mouseleave', () => start());
	});

	setActive(0);
	if (!window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
		start();
	}
})();
