<script>
	import { onMount } from "svelte";
	let cells = Array(225)
		.fill()
		.map(() => ({ char: "", isBlock: true }));
	let variables = [],
		status = "Connecting...",
		isConnected = false,
		socket;
	let frameRequested = false;

	onMount(() => connect());

	function connect() {
		const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
		socket = new WebSocket(`${protocol}//${location.host}/ws`);

		socket.onopen = () => {
			isConnected = true;
			status = "Ready";
		};
		socket.onmessage = (e) => handleEvent(JSON.parse(e.data));
		socket.onclose = () => {
			isConnected = false;
			setTimeout(connect, 2000);
		};
	}
	async function start() {
		status = "Generating...";
		const res = await fetch("/api/solve", { method: "POST" });
		const data = await res.json();

		console.log("Received variables:", data); // Check your browser console!
		variables = data;

		// Create a fresh grid state
		const newCells = Array(225)
			.fill()
			.map((_, i) => ({
				char: "",
				isBlock: true,
			}));

		// Mark cells as white if they belong to a word
		variables.forEach((v) => {
			v.indices.forEach((idx) => {
				if (newCells[idx]) {
					newCells[idx].isBlock = false;
				}
			});
		});

		cells = newCells;
		status = "Solving...";
	}

	function handleEvent(data) {
		if (data.type === "PLACE" || data.type === "BACK") {
			const v = variables.find((v) => v.id === data.id);
			if (!v) {
				console.error("Could not find variable with ID:", data.id);
				return;
			}

			v.indices.forEach((idx, i) => {
				// data.word is the full word from the solver
				// We take the character at the correct offset
				const char = data.word[i] || "";
				cells[idx].char = char;
			});

			cells = [...cells]; // Force Svelte to refresh
		}
	}
</script>

<main>
	<h1>Crossword Engine</h1>
	<div class="status">{status}</div>
	<button on:click={start} disabled={!isConnected}>Generate & Solve</button>

	<div class="grid">
		{#each cells as cell, i (i)}
			<div class="cell" class:block={cell.isBlock}>{cell.char}</div>
		{/each}
	</div>
</main>

<style>
	:global(body) {
		background: #121212;
		color: #eee;
		font-family: sans-serif;
		display: grid;
		place-items: center;
		min-height: 100vh;
		margin: 0;
	}
	.status {
		margin: 10px 0;
		color: #888;
	}
	button {
		background: #646cff;
		border: none;
		padding: 10px 20px;
		border-radius: 5px;
		color: white;
		cursor: pointer;
	}
	button:disabled {
		background: #333;
	}
	.grid {
		display: grid;
		grid-template-columns: repeat(15, 30px);
		gap: 1px;
		background: #333;
		padding: 5px;
		margin-top: 20px;
	}
	.cell {
		width: 30px;
		height: 30px;
		background: white;
		color: black;
		display: grid;
		place-items: center;
		font-weight: bold;
		text-transform: uppercase;
	}
	.cell.block {
		background: #121212;
	}
</style>
