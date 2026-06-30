package ict_mikrotik

/* ======================= ict_router_status
  ONLINE | OFFLINE | UNREACHABLE | UNKNOWN

========================== ict_firewall_chain
  input | output | forward

========================== ict_firewall_action
  accept | drop | log |
  add_dst_to_address_list |
  add_src_to_address_list |
  fasttrack_connection |
  passthrough | jump |
  reject | return | tarpit
========================== */

/* ======================= ict_mikrotik_router
  id                       String            @id @default(uuid())
  name                     String
  ip_address               String            @unique
  api_port                 Int               @default(8728)
  username                 String
  password                 String
  status                   ict_router_status @default(UNKNOWN)
  last_seen                DateTime?
  created_at               DateTime          @default(now())
  updated_at               DateTime          @default(now())
  metrics                  ict_mikrotik_metric[]
  interfaces               ict_mikrotik_interface[]
  filter_rules             ict_mikrotik_filter_rule[]
  address_lists            ict_mikrotik_address_list[]
========================== */

/* ======================= ict_mikrotik_interface
  id                       String            @id @default(uuid())
  router_id                String
  name                     String
  type                     String
  mac_address              String?
  description              String?
  created_at               DateTime          @default(now())
  updated_at               DateTime          @default(now())
  router                   ict_mikrotik_router @relation(fields: [router_id], references: [id], onDelete: Cascade)
  metrics                  ict_mikrotik_intmet[]

  @@unique([router_id, name])
========================== */

/* ======================= ict_mikrotik_metric
  id                       String            @id @default(uuid())
  router_id                String
  cpu_load                 Int
  free_memory              BigInt
  uptime                   String
  timestamp                DateTime          @default(now())
  router                   ict_mikrotik_router @relation(fields: [router_id], references: [id], onDelete: Cascade)

  @@index([router_id, timestamp])
========================== */

/* ======================= ict_mikrotik_intmet
  id                       String            @id @default(uuid())
  interface_id             String
  tx_bytes                 BigInt
  rx_bytes                 BigInt
  tx_packets               BigInt
  rx_packets               BigInt
  timestamp                DateTime          @default(now())
  interface                ict_mikrotik_interface @relation(fields: [interface_id], references: [id], onDelete: Cascade)

  @@index([interface_id, timestamp])
========================== */

/* ======================= ict_mikrotik_filter_rule
  id                       String            @id @default(uuid())
  router_id                String
  mikrotik_id              String?
  chain                    ict_firewall_chain
  action                   ict_firewall_action
  protocol                 String?
  src_port                 String?
  dst_port                 String?
  src_address              String?
  dst_address              String?
  comment                  String?
  disabled                 Boolean           @default(false)
  position                 Int
  created_at               DateTime          @default(now())
  updated_at               DateTime          @default(now())
  router                   ict_mikrotik_router @relation(fields: [router_id], references: [id], onDelete: Cascade)

  @@unique([router_id, mikrotik_id])
  @@index([router_id, position])
========================== */

/* ======================= ict_mikrotik_address_list
  id                       String            @id @default(uuid())
  router_id                String
  mikrotik_id              String?
  list_name                String
  address                  String
  comment                  String?
  timeout                  String?
  disabled                 Boolean           @default(false)
  created_at               DateTime          @default(now())
  updated_at               DateTime          @default(now())
  router                   ict_mikrotik_router @relation(fields: [router_id], references: [id], onDelete: Cascade)

  @@unique([router_id, mikrotik_id])
  @@index([router_id, list_name])
========================== */

type Repository interface {
}

type UseCase interface {
}
