<template>
	<div class="container">
		<div class="jumbotron">
			<h1>{{ error }}</h1>
			<p>feelsbadman</p>
		</div>
	</div>
</template>

<script>
	export default {
		name: "Error",
		props: {
			error: {
				type: String,
				default: "Page Not Found",
			},
		},
	}
</script>

<style scoped>

</style>
