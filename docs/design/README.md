# Design documents

As part of the design process for a new feature in terraform-plugin-framework, the Terraform Plugin SDK Team may create a design document, whose purpose is to:

 - outline the options for the new feature,
 - describe each option's tradeoffs,
 - evaluate each option against our design principles,
 - provide context and relevant background information needed to understand the above,
 - record which option we decide to take.

Design documents are also created for fundamental design positions, such as the use of interface types over struct types for providers and resources.

The core design principles we use to judge designs are:

1. How unit testable is it? 

2. How much room do we have to change while maintaining compatibility?

3. How discoverable is it to new users?

4. How verbose is it?

5. How Go-native does it feel?

The design documents in this folder are artefacts of the design process, and are usually not updated after implementation begins. They are therefore best consulted as an aid to understanding the decisions behind a given feature's design, while questions about its API and behaviour are better answered by up-to-date documentation.
