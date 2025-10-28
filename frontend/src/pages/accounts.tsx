import { useState, useMemo } from 'react'
import { useForm, type Resolver } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Plus, Trash2, ExternalLink, Loader2, CheckCircle2, XCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Dialog } from '@/components/ui/dialog'
import {
  Form,
  FormField,
  FormItem,
  FormLabel,
  FormControl,
  FormMessage,
} from '@/components/ui/form'
import {
  useAccounts,
  useCreateAccount,
  useCompleteAccount,
  useDeleteAccount,
} from '@/hooks/use-accounts'

// Validation schemas
const baseStep1Schema = z.object({
  appName: z
    .string()
    .min(1, 'App name is required')
    .min(3, 'App name must be at least 3 characters'),
  orgId: z.string().optional(),
})

function createStep1SchemaWithUniqueCheck(
  existingAccounts: Array<{ id: string; name: string; organizationUuid: string }>
) {
  return baseStep1Schema
    .refine(
      (data) => {
        const isDuplicate = existingAccounts.some(
          (a) => a.name.toLowerCase() === data.appName.toLowerCase()
        )
        return !isDuplicate
      },
      {
        message: 'Account name must be unique',
        path: ['appName'],
      }
    )
    .refine(
      (data) => {
        if (!data.orgId) return true
        const isDuplicate = existingAccounts.some((a) => a.organizationUuid === data.orgId)
        return !isDuplicate
      },
      {
        message: 'Organization ID must be unique',
        path: ['orgId'],
      }
    )
}

const step2Schema = z.object({
  authCode: z
    .string()
    .min(1, 'Authorization code is required')
    .min(10, 'Authorization code appears to be invalid'),
})

type Step1FormData = z.infer<typeof baseStep1Schema>
type Step2FormData = z.infer<typeof step2Schema>

export function AccountsPage() {
  // React Query hooks
  const { data: accounts = [], isLoading } = useAccounts()
  const createMutation = useCreateAccount()
  const completeMutation = useCompleteAccount()
  const deleteMutation = useDeleteAccount()

  // Modal state
  const [showModal, setShowModal] = useState(false)
  const [step, setStep] = useState(1)
  const [authUrl, setAuthUrl] = useState('')
  const [state, setState] = useState('')
  const [codeVerifier, setCodeVerifier] = useState('')
  const [appNameForStep2, setAppNameForStep2] = useState('')
  const [accountToDelete, setAccountToDelete] = useState<string | null>(null)

  // Form handlers for each step
  const form1 = useForm<Step1FormData>({
    resolver: zodResolver(
      useMemo(() => createStep1SchemaWithUniqueCheck(accounts), [accounts])
    ) as Resolver<Step1FormData>,
    defaultValues: {
      appName: '',
      orgId: '',
    },
  })

  const form2 = useForm<Step2FormData>({
    resolver: zodResolver(step2Schema),
    defaultValues: {
      authCode: '',
    },
  })

  // Step 1: Start OAuth flow
  const handleStartOAuth = async (data: Step1FormData) => {
    try {
      const result = await createMutation.mutateAsync({
        name: data.appName,
        orgId: data.orgId || undefined,
      })
      setAuthUrl(result.authorizationUrl)
      setState(result.state)
      setCodeVerifier(result.codeVerifier)
      setAppNameForStep2(data.appName)
      setStep(2)
    } catch (err) {
      form1.setError('root', {
        type: 'server',
        message: err instanceof Error ? err.message : 'Failed to start OAuth',
      })
    }
  }

  // Step 2: Complete OAuth flow
  const handleCompleteOAuth = async (data: Step2FormData) => {
    try {
      await completeMutation.mutateAsync({
        name: appNameForStep2,
        code: data.authCode,
        state,
        codeVerifier,
        orgId: form1.getValues('orgId') || undefined,
      })
      resetModal()
    } catch (err) {
      form2.setError('root', {
        type: 'server',
        message: err instanceof Error ? err.message : 'Failed to complete OAuth',
      })
    }
  }

  // Delete account
  const handleDeleteClick = (id: string) => {
    setAccountToDelete(id)
  }

  const handleConfirmDelete = async () => {
    if (!accountToDelete) return
    try {
      await deleteMutation.mutateAsync(accountToDelete)
      setAccountToDelete(null)
    } catch (error) {
      console.error('Failed to delete account:', error)
    }
  }

  const resetModal = () => {
    setShowModal(false)
    setStep(1)
    form1.reset()
    form2.reset()
    setAuthUrl('')
    setState('')
    setCodeVerifier('')
    setAppNameForStep2('')
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-foreground text-3xl font-bold tracking-tight">Accounts (OAuth)</h1>
          <p className="text-muted-foreground">Manage Claude OAuth accounts</p>
        </div>
        <Button onClick={() => setShowModal(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Create Account
        </Button>
      </div>

      {/* Accounts Table */}
      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="py-12 text-center">
              <Loader2 className="text-primary mx-auto h-8 w-8 animate-spin" />
              <p className="text-muted-foreground mt-2">Loading accounts...</p>
            </div>
          ) : accounts.length === 0 ? (
            <div className="py-12 text-center">
              <p className="text-muted-foreground">No accounts yet. Create one to get started!</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Organization UUID</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Expires At</TableHead>
                  <TableHead>Created At</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {accounts.map((account) => (
                  <TableRow key={account.id}>
                    <TableCell className="font-medium">{account.name}</TableCell>
                    <TableCell>
                      <code className="bg-muted text-foreground rounded px-2 py-1 text-xs">
                        {account.organizationUuid}
                      </code>
                    </TableCell>
                    <TableCell>
                      <span
                        className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
                          account.status === 'active'
                            ? 'bg-green-500/10 text-green-500'
                            : 'bg-gray-500/10 text-gray-500'
                        }`}
                      >
                        {account.status}
                      </span>
                    </TableCell>
                    <TableCell className="text-foreground text-sm">
                      {new Date(account.expiresAt).toLocaleString()}
                    </TableCell>
                    <TableCell className="text-foreground/70 text-sm">
                      {new Date(account.createdAt).toLocaleString()}
                    </TableCell>
                    <TableCell className="text-right">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleDeleteClick(account.id)}
                        disabled={deleteMutation.isPending}
                      >
                        {deleteMutation.isPending ? (
                          <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                          <Trash2 className="text-destructive h-4 w-4" />
                        )}
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* Create Modal */}
      {showModal && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
          onClick={(e) => {
            if (e.target === e.currentTarget) {
              resetModal()
            }
          }}
        >
          <Card className="w-full max-w-lg">
            <CardHeader>
              <CardTitle>Create Account</CardTitle>
              <CardDescription>Step {step} of 2</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {step === 1 && (
                <Form {...form1}>
                  <form onSubmit={form1.handleSubmit(handleStartOAuth)} className="space-y-4">
                    <FormField
                      control={form1.control}
                      name="appName"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Account Name *</FormLabel>
                          <FormControl>
                            <Input
                              placeholder="My Account"
                              disabled={createMutation.isPending}
                              {...field}
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />

                    <FormField
                      control={form1.control}
                      name="orgId"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Organization ID (Optional)</FormLabel>
                          <FormControl>
                            <Input
                              placeholder="org_..."
                              disabled={createMutation.isPending}
                              {...field}
                            />
                          </FormControl>
                          <p className="text-muted-foreground mt-1 text-xs">
                            If not provided, will be fetched from your account
                          </p>
                          <FormMessage />
                        </FormItem>
                      )}
                    />

                    {form1.formState.errors.root && (
                      <Alert variant="destructive">
                        <XCircle className="h-4 w-4" />
                        <AlertDescription>{form1.formState.errors.root.message}</AlertDescription>
                      </Alert>
                    )}

                    <div className="flex gap-2">
                      <Button
                        type="button"
                        variant="outline"
                        onClick={resetModal}
                        disabled={createMutation.isPending}
                        className="flex-1"
                      >
                        Cancel
                      </Button>
                      <Button type="submit" disabled={createMutation.isPending} className="flex-1">
                        {createMutation.isPending ? (
                          <>
                            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                            Generating...
                          </>
                        ) : (
                          'Authenticate with Claude'
                        )}
                      </Button>
                    </div>
                  </form>
                </Form>
              )}

              {step === 2 && (
                <Form {...form2}>
                  <form onSubmit={form2.handleSubmit(handleCompleteOAuth)} className="space-y-4">
                    <Alert>
                      <CheckCircle2 className="h-4 w-4" />
                      <AlertDescription>
                        <strong>Step 1:</strong> Click the button below to authorize with Claude
                      </AlertDescription>
                    </Alert>

                    <Button asChild className="w-full">
                      <a href={authUrl} target="_blank" rel="noopener noreferrer">
                        Open Claude Authorization
                        <ExternalLink className="ml-2 h-4 w-4" />
                      </a>
                    </Button>

                    <div className="border-t pt-4">
                      <FormField
                        control={form2.control}
                        name="authCode"
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>
                              <strong>Step 2:</strong> Paste Authorization Code
                            </FormLabel>
                            <FormControl>
                              <Input
                                placeholder="Paste the code from the callback URL"
                                disabled={completeMutation.isPending}
                                {...field}
                              />
                            </FormControl>
                            <p className="text-muted-foreground mt-1 text-xs">
                              After authorizing, copy the <code>code</code> parameter from the
                              redirect URL
                            </p>
                            <FormMessage />
                          </FormItem>
                        )}
                      />
                    </div>

                    {form2.formState.errors.root && (
                      <Alert variant="destructive">
                        <XCircle className="h-4 w-4" />
                        <AlertDescription>{form2.formState.errors.root.message}</AlertDescription>
                      </Alert>
                    )}

                    <div className="flex gap-2">
                      <Button
                        type="button"
                        variant="outline"
                        onClick={resetModal}
                        disabled={completeMutation.isPending}
                        className="flex-1"
                      >
                        Cancel
                      </Button>
                      <Button
                        type="submit"
                        disabled={completeMutation.isPending}
                        className="flex-1"
                      >
                        {completeMutation.isPending ? (
                          <>
                            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                            Completing...
                          </>
                        ) : (
                          'Complete Setup'
                        )}
                      </Button>
                    </div>
                  </form>
                </Form>
              )}
            </CardContent>
          </Card>
        </div>
      )}

      <Dialog
        open={!!accountToDelete}
        onClose={() => setAccountToDelete(null)}
        title="Delete Account?"
      >
        <p className="text-muted-foreground mb-4 text-sm">
          This action cannot be undone. This will permanently delete the account and all associated
          data.
        </p>
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => setAccountToDelete(null)}>
            Cancel
          </Button>
          <Button onClick={handleConfirmDelete} variant="destructive">
            {deleteMutation.isPending ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Deleting...
              </>
            ) : (
              'Delete Account'
            )}
          </Button>
        </div>
      </Dialog>
    </div>
  )
}
