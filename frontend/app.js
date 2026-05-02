window.addEventListener("load", () => {
	const tg = window.Telegram?.WebApp;
	if (!tg) {
		document.getElementById("content").textContent =
			"Приложение открыто без Telegram.";
		return;
	}

	tg.ready();
	try {
		tg.expand?.();
	} catch {
		// ignore Telegram client quirks across versions
	}

	const prefersDark =
		window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches;

	if (tg.colorScheme === "dark" || (tg.colorScheme !== "light" && prefersDark)) {
		document.documentElement.style.colorScheme = "dark";
	}

	document.getElementById("content").textContent =
		"Список помещений появится здесь после подключения сервера.";
});
