// Sequential playback of the lifecycle scenario animations.
// One row is .active at a time. The play/pause button toggles the rotation
// and freezes the active scenario where it is.
(() => {
	const list = document.querySelector('.anim-list');
	if (!list) return;
	const figures = list.querySelectorAll('.anim');
	const btn = document.querySelector('.play-pause');
	if (!figures.length || !btn) return;

	const cycleMs = 10000;
	let current = 0;
	let timer = null;
	let playing = true;

	const setActive = (i) => {
		current = ((i % figures.length) + figures.length) % figures.length;
		figures.forEach((el, idx) => el.classList.toggle('active', idx === current));
	};

	const advance = () => setActive(current + 1);

	const setButton = () => {
		btn.textContent = playing ? 'Pause' : 'Play';
		btn.setAttribute('aria-pressed', playing ? 'false' : 'true');
	};

	const play = () => {
		if (playing) return;
		playing = true;
		list.classList.remove('paused');
		clearInterval(timer);
		timer = setInterval(advance, cycleMs);
		setButton();
	};

	const pause = () => {
		if (!playing) return;
		playing = false;
		clearInterval(timer);
		list.classList.add('paused');
		setButton();
	};

	btn.addEventListener('click', () => (playing ? pause() : play()));

	setActive(0);
	if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
		pause();
	} else {
		timer = setInterval(advance, cycleMs);
		setButton();
	}
})();
