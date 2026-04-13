<script>
	import { Card, Button } from 'flowbite-svelte';
	import { HeartSolid } from 'flowbite-svelte-icons';
</script>

<Card
	size="none"
	class="border-2 border-dashed border-gray-300 bg-transparent p-6 text-center dark:border-gray-600"
>
	<HeartSolid class="mx-auto mb-2 text-pink-500" size="lg" />
	<h3 class="text-lg font-bold dark:text-white">Find this tool useful?</h3>
	<p class="mb-4 text-sm text-gray-500">Help keep the servers running and the EFL data updated.</p>
	<div class="flex justify-center gap-4">
		<Button color="alternative" href="https://github.com/sponsors/cloudkucooland" target="_blank"
			>GitHub Sponsor</Button
		>
		<Button color="blue" href="https://venmo.com/@julysun-store" target="_blank">Venmo</Button>
	</div>
</Card>
