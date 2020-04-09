<template>
    <div>
        <table class="table">
            <thead class="thead-dark">
                <th id="id" scope="col">#</th>
                <th scope="col">Name</th>
                <th scope="col">Author</th>
                <th scope="col">From</th>
                <th scope="col">To</th>
            </thead>
            <tbody>
            <tr v-for="prs in prs">
                <th scope="row">{{ prs.id }}</th>
                <td>
                    <router-link :to="`/${$route.params.user}/${$route.params.project}/pr/${prs.id}`">
                        {{ prs.title }}
                    </router-link>
                </td>
                <td>
                    <router-link :to="`/${prs.id}`">
                        @{{ prs.author }}
                    </router-link>
                </td>
                <td>
                    {{ prs.from }}
                </td>
                <td>
                    {{ prs.to }}
                </td>
            </tr>
            <tr>
                <td>
                    <router-link class="btn btn-secondary btn-lg" :to="`/${$route.params.user}/${$route.params.project}/pr/new`">
                        add pull request
                    </router-link>
                </td>
            </tr>
            </tbody>
        </table>
    </div>
</template>

<script>
    export default {
        name: "PrsList",
        data() {
            return {
                prs: [],
            }
        },
        async created() {
            const { user, project } = this.$route.params
            const res = await fetch(`/api/v1/user/${user}/project/${project}/prs`)

            if (res.ok) {
                this.prs = await res.json()
            } else {
                console.error(res.statusText)
            }
        },
    }
</script>

<style scoped>
    th{
        width:250px;
    }

    #id{
        width:50px
    }
</style>
