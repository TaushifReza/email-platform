## 25. Compliance considerations

The product is **general-purpose** and does not restrict content by recipient relationship (existing contacts, cold or purchased lists). **The tenant is responsible** for the legality of what it sends and for its provider's policies. The platform provides tools, not legal assurance. This section is design guidance, not legal advice. (C)

- **COM-01 [M·C]** Design to the **strictest common baseline** of privacy and email-marketing rules. Current tenants are in Nepal. Applicable local law and any foreign law relevant to recipients must be reviewed by counsel before commercialization.
- **COM-02 [M·C]** **Unsubscribe and consent tools:** marketing campaigns require an unsubscribe link (hard block) and should include the sender's postal address (warning). Opt-outs take effect immediately. Sender declaration at launch.
- **COM-03 [M·C]** **Privacy-by-design** principles are required: data minimization (no permanent contact database), purpose limitation (data tied to a campaign), storage limitation (retention purge), deletion on request, security, and access logging.
- **COM-04 [M·C]** **Roles under privacy law:** tenant is the data controller, platform is the processor. A **Terms of Service** (acceptable use), **privacy notice** and **data processing statement** are required before commercialization and are **post-MVP** (OD-05).
- **COM-05 [M·C]** **Provider and mailbox-provider policies:** tenants must respect their provider's terms. Large senders are increasingly required by major mailbox providers to authenticate with SPF, DKIM and DMARC and to offer one-click unsubscribe. The platform assists through PRV-08 and UNS-01 but does not guarantee deliverability.
- **COM-06 [M·C]** **Tenant-supplied content** and data are the tenant's responsibility. The platform provides abuse reporting, suspension and investigation tools.
- **COM-07 [M·C]** **Backups:** deleted data may persist in encrypted backups for up to about **30 days**. After any restore from backup, previous deletions and purges MUST be re-applied before the system returns to service. (Re-apply rule A.)
- **COM-08 [M·C]** **Transactional email** is not required to carry an unsubscribe link, but organizations should not use the transactional type to avoid marketing rules. Type selection is the sender's declaration.
