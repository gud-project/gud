<template>
	<div class="container">
		<form class="jumbotron" @submit="send">
			<p style="color: red" v-if="error">{{ error }}</p>
			<p v-if="pr" class="container">
				<label><select v-model="info.from" required class="custom-select">
					<option disabled selected>Source branch</option>
					<option v-for="branch in branches">{{ branch }}</option>
				</select></label>
				⇒
				<label><select v-model="info.to" required class="custom-select">
					<option disabled selected>Target branch</option>
					<option v-for="branch in branches">{{ branch }}</option>
				</select></label>
			</p>

			<p><label>
				<input type="text" placeholder="Title" v-model="info.title" required class="form-control"/>
			</label></p>
			<p><label>
				<textarea placeholder="Description" v-model="info.content" class="form-control" rows="5" cols="35"></textarea>
			</label></p>

			<input class="btn btn-primary btn-lg" type="submit" value="Submit" />
		</form>
	</div>
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
				this.branches = Object.keys(await res.json())
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
	select{
		width:130px
	}
</style>
