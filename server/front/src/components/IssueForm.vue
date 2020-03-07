<template>
	<form @submit="send">
		<p style="color: red" v-if="error">{{ error }}</p>
		<p v-if="pr">
			<label><select v-model="info.from" required>
				<option disabled selected value="">From</option>
				<option v-for="branch in branches">{{ branch }}</option>
			</select></label>
			⇒
			<label><select v-model="info.from" required>
				<option disabled selected value="">To</option>
				<option v-for="branch in branches">{{ branch }}</option>
			</select></label>
		</p>
		<p><label>
			<input type="text" placeholder="Title" v-model="info.title" required />
		</label></p>
		<p><label>
			<textarea placeholder="Description" v-model="info.content" />
		</label></p>
		
		<input type="submit" value="Submit" />
	</form>
</template>

<script>
	export default {
		name: "IssueForm",
		props: {
			pr: Boolean,
		},
		data() {
			return {
				info: {
					title: null,
					content: "",
					from: null,
					to: null,
				},
				branches: [],
				error: null,
			}
		},
		async created() {
			const { user, project } = this.$route.params
			const res = await fetch(`/api/v1/user/${user}/project/${project}/branches`)
			
			if (res.ok) {
				this.branches = await res.json()
			} else {
				console.error(res.statusText)
			}
		},
		methods: {
			async send(e) {
				e.preventDefault()
				
				if (this.pr && this.info.from === this.info.to) {
					this.error = "cannot merge a branch with itself."
					return false
				}
				
				const { user, project } = this.$route.params
				const category = this.pr ? 'pr' : 'issue'
				const res = await fetch(`/api/v1/user/${user}/project/${project}/${category}s/create`, {
					method: "POST",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify(this.info),
				})
				const data = await res.json()
				
				if (res.ok) {
					await this.$router.push(`/${user}/${project}/${category}/${data.id}`)
				} else {
					this.error = data.error
				}
			}
		}
	}
</script>

<style scoped>

</style>
