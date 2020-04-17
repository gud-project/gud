<template>
    <div class="container">
        <div class="container">
            <div class="jumbotron">
                <h1>What is Gud?</h1>
                <p>
                    Gud is a version control system, attempting to answer the problems coming up with the<br />
                    existing version controls systems, and add new, modern features to them.<br />
                    Some of the new major features are an undo command in the CLI,<br />
                    and automatic update.
                </p>
            </div>
        </div>
        <div class="container">
            <div class="jumbotron">
                <h1>How to get it</h1>
                <p>
                    Just press the download button. Gud is cross-platform
                    and works on Linux, MacOS and Windows.<br />
                </p>
                <form :action="url"><button type="submit" class="btn btn-lg btn-primary">Download</button></form>
                <p>
                    <br />In order the use the "gud" root command in any file with your terminal,<br />
                    add the path to your "gud" executable to your PATH environment variable.
                </p>
            </div>
        </div>
    </div>
</template>

<script>
    export default {
        name: "HomeForm",
        computed: {
            url() {
                if (navigator.platform.includes("Win")) {
                    return "/release/windows/gud.exe"
                }
                if (navigator.platform.includes("Mac")) {
                    return "/release/darwin/gud"
                }
                return "/release/linux/gud"
            }
        }
    }
</script>

<style scoped>

</style>
