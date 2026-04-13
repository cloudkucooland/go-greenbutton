<script>
	import { onMount } from 'svelte';
	import { Navbar, NavBrand, NavLi, NavUl, Fileupload, Spinner } from 'flowbite-svelte';
	import { InfoCircleSolid } from 'flowbite-svelte-icons';
	import PlanTable from './lib/PlanTable.svelte';
	import Conditions from './lib/Conditions.svelte';
	import Support from './lib/Support.svelte';

	let plans = $state([]);
	let results = $state(null);
	let loading = $state(false);
	let error = $state('');

	let selectedFiles = $state(null);

	onMount(async () => {
		try {
			const res = await fetch('http://chorus.local:8090/plans');
			plans = await res.json();
		} catch (e) {
			error = 'Failed to load plans from server.';
		}
	});

	async function handleUpload(event) {
		loading = true;
		error = '';

		const sf = selectedFiles;
		const formData = new FormData();
		formData.append('d', sf[0]);

		try {
			const response = await fetch('http://chorus.local:8090/upload', {
				method: 'POST',
				body: formData
			});

			if (!response.ok) throw new Error(await response.text());
			results = await response.json();
		} catch (e) {
			error = e.message;
		} finally {
			loading = false;
		}
	}
</script>

<Navbar>
	<NavBrand href="/">
		<span class="self-center text-xl font-semibold whitespace-nowrap dark:text-white">
			Texas Solar Plan Simulator
		</span>
	</NavBrand>
	<NavUl>
		<NavLi href="https://github.com/cloudkucooland/go-greenbutton" target="_blank">GitHub</NavLi>
	</NavUl>
</Navbar>

<main class="container mx-auto flex max-w-4xl flex-col space-y-8 p-4">
	<section id="conditions">
		<Conditions />
	</section>


	<section id="upload" class="rounded-lg bg-white p-6 shadow-md dark:bg-gray-800">
		<h2 class="mb-4 text-lg font-bold dark:text-white">1. Upload SMT Data</h2>

		<div
			class="mb-6 rounded-lg border border-blue-100 bg-blue-50 p-4 text-sm text-blue-800 dark:border-blue-800 dark:bg-gray-700 dark:text-blue-400"
		>
			<div class="mb-2 flex items-center gap-2 font-bold tracking-tight uppercase">
				<InfoCircleSolid size="sm" />
				How to get your data
			</div>
			<ol class="list-inside list-decimal space-y-2">
				<li>
					Log in to <a
						href="https://www.smartmetertexas.com"
						target="_blank"
						class="font-semibold underline hover:text-blue-600">Smart Meter Texas (SMT)</a
					>.
				</li>
				<li>Change the <b>"Start Date"</b> to 2 years in the past (SMT limit).</li>
				<li>Ensure that <b>"Report Type"</b> is set to <b>"Energy Data 15 Min Interval"</b>.</li>
				<li>Click <b>"Submit Update"</b>.</li>
				<li>Click <b>"Export My Report"</b> to download your CSV.</li>
				<li>Once the file downloads, upload it here.</li>
			</ol>
		</div>

		<p class="mb-4 text-sm text-gray-500">Upload your Smart Meter Texas CSV.</p>

		<Fileupload
			id="smt-upload"
			class="mb-2"
			bind:files={selectedFiles}
			onchange={handleUpload}
			disabled={loading}
		/>

		{#if loading}
			<div class="mt-4 flex items-center space-x-2">
				<Spinner size="4" />
				<span class="text-sm text-blue-600">Calculating solar scenarios...</span>
			</div>
		{/if}

		{#if error}
			<div class="mt-4 rounded-md bg-red-100 p-3 text-sm text-red-700">
				Error: {error}
			</div>
		{/if}
	</section>

	{#if results}
		<section class="space-y-4">
			<h2 class="text-lg font-bold dark:text-white">2. Simulation Results</h2>
			<PlanTable {results} {plans} />
		</section>
	{/if}
	<section class="space-y-4">
		<Support />
	</section>
</main>
