# GPSR

**G**roups API + **P**eople Web Service + **S**lurm = **R**eport

Brought to you by the **UW-IT Research Computing Facilitation Team**.

This repository contains the code for the GPSR tool, which integrates the Groups API and People Web Service as internal UW-IT tools with Slurm accounting data to generate reports for university leadership.

The certificates are hard-coded for `klone-head01` but you're welcome to use your own x.509 certificate and modify the code accordingly.

Use the standard `go build` then run the `gpsr` binary to generate a CSV format to STDOUT. Redirect the output for further processing.
